package embind

import (
	"context"
	"fmt"
	"reflect"

	"github.com/tetratelabs/wazero/api"
)

type classProperty struct {
	name         string
	enumerable   bool
	configurable bool
	readOnly     bool
	setterType   registeredType
	getterType   registeredType
	set          func(ctx context.Context, this any, v any) error
	get          func(ctx context.Context, this any) (any, error)
}

func (cp *classProperty) Name() string {
	return cp.name
}

func (cp *classProperty) GetterType() IType {
	return &exposedType{
		registeredType: cp.getterType,
	}
}

func (cp *classProperty) SetterType() IType {
	return &exposedType{
		registeredType: cp.setterType,
	}
}

func (cp *classProperty) ReadOnly() bool {
	return cp.readOnly
}

type classConstructor struct {
	fn            publicSymbolFn
	argumentTypes []registeredType
	resultType    registeredType
}

type classType struct {
	baseType
	legalFunctionName    string
	baseClass            *classType
	rawDestructor        api.Function
	getActualType        api.Function
	upcast               api.Function
	downcast             api.Function
	derivedClasses       []*classType
	goStruct             any
	hasGoStruct          bool
	hasCppClass          bool
	pureVirtualFunctions []string
	methods              map[string]*publicSymbol
	properties           map[string]*classProperty
	constructors         map[int32]*classConstructor
}

func (erc *classType) FromWireType(ctx context.Context, mod api.Module, value uint64) (any, error) {
	panic("FromWireType should not be called on classes")
}

func (erc *classType) ToWireType(ctx context.Context, mod api.Module, destructors *[]*destructorFunc, o any) (uint64, error) {
	panic("ToWireType should not be called on classes")
}

func (erc *classType) ReadValueFromPointer(ctx context.Context, mod api.Module, pointer uint32) (any, error) {
	panic("ReadValueFromPointer should not be called on classes")
}

func (erc *classType) GoType() string {
	return erc.legalFunctionName
}

func (erc *classType) validate() error {
	if !erc.hasGoStruct || !erc.hasCppClass {
		return nil
	}

	// @todo: implement validator here.
	// @todo: we want to check if the Go struct implements everything we need.
	//log.Printf("Running validator on %T", erc.goStruct)

	//for i := range erc.constructors {
	//	log.Println(erc.constructors[i].argTypes)
	//	log.Println(erc.constructors[i].resultType)
	//}

	//log.Println(erc.constructors)
	///log.Println(erc.methods)
	//log.Println(erc.properties)

	return nil
}

func (erc *classType) isDeleted(handle IClassBase) bool {
	return handle.getRegisteredPtrTypeRecord().ptr == 0
}

func (erc *classType) deleteLater(handle IClassBase) (any, error) {
	registeredPtrTypeRecord := handle.getRegisteredPtrTypeRecord()
	if registeredPtrTypeRecord.ptr == 0 {
		return nil, fmt.Errorf("class handle already deleted")
	}

	if registeredPtrTypeRecord.deleteScheduled && !registeredPtrTypeRecord.preservePointerOnDelete {
		return nil, fmt.Errorf("object already scheduled for deletion")
	}

	// @todo: implement me.
	/*
	   deletionQueue.push(this);
	   if (deletionQueue.length === 1 && delayFunction) {
	     delayFunction(flushPendingDeletes);
	   }
	*/

	registeredPtrTypeRecord.deleteScheduled = true

	return handle, nil
}

func (erc *classType) isAliasOf(ctx context.Context, first, second IClassBase) (bool, error) {
	leftClass := first.getRegisteredPtrTypeRecord().ptrType.registeredClass
	left := first.getRegisteredPtrTypeRecord().ptr
	rightClass := second.getRegisteredPtrTypeRecord().ptrType.registeredClass
	right := second.getRegisteredPtrTypeRecord().ptr

	for leftClass.baseClass != nil {
		leftRes, err := leftClass.upcast.Call(ctx, api.EncodeU32(left))
		if err != nil {
			return false, err
		}
		left = api.DecodeU32(leftRes[0])
		leftClass = leftClass.baseClass
	}

	for rightClass.baseClass != nil {
		rightRes, err := rightClass.upcast.Call(ctx, api.EncodeU32(right))
		if err != nil {
			return false, err
		}
		right = api.DecodeU32(rightRes[0])
		rightClass = rightClass.baseClass
	}

	return leftClass == rightClass && left == right, nil
}

func (erc *classType) clone(ctx context.Context, from IClassBase) (IClassBase, error) {
	registeredPtrTypeRecord := from.getRegisteredPtrTypeRecord()
	if registeredPtrTypeRecord.ptr == 0 {
		return nil, fmt.Errorf("class handle already deleted")
	}

	if registeredPtrTypeRecord.preservePointerOnDelete {
		registeredPtrTypeRecord.count.value += 1
		return from, nil
	}

	clone, err := erc.getNewInstance(ctx, registeredPtrTypeRecord.shallowCopyInternalPointer())
	if err != nil {
		return nil, err
	}

	clone.getRegisteredPtrTypeRecord().count.value += 1
	clone.getRegisteredPtrTypeRecord().deleteScheduled = false
	return clone, nil
}

func (erc *classType) delete(ctx context.Context, handle IClassBase) error {
	registeredPtrTypeRecord := handle.getRegisteredPtrTypeRecord()
	if registeredPtrTypeRecord.ptr == 0 {
		return fmt.Errorf("class handle already deleted")
	}

	if registeredPtrTypeRecord.deleteScheduled && !registeredPtrTypeRecord.preservePointerOnDelete {
		return fmt.Errorf("object already scheduled for deletion")
	}

	// @TODO: We don't use the finalizer anymore. When we use the Go GC finalizer
	// 		  we should properly set something on the record to let the finalizer
	// 		  know it shouldn't do anything.
	//
	// err := registeredPtrTypeRecord.detachFinalizer(ctx)
	// if err != nil {
	// 	return err
	// }

	err := registeredPtrTypeRecord.releaseClassHandle(ctx)
	if err != nil {
		return err
	}

	if registeredPtrTypeRecord.preservePointerOnDelete {
		registeredPtrTypeRecord.smartPtr = 0
		registeredPtrTypeRecord.ptr = 0
	}

	return nil
}

func (erc *classType) getNewInstance(ctx context.Context, record *registeredPointerTypeRecord) (IClassBase, error) {
	e := MustGetEngineFromContext(ctx, nil).(*engine)
	classBase := &ClassBase{
		classType:               erc,
		ptr:                     record.ptr,
		ptrType:                 record.ptrType,
		registeredPtrTypeRecord: record,
		engine:                  e,
	}

	// If we have a Go struct, wrap the resulting class in it.
	if erc.hasGoStruct {
		typeElem := reflect.TypeOf(erc.goStruct).Elem()
		newElem := reflect.New(typeElem)
		f := newElem.Elem().FieldByName("ClassBase")
		if f.IsValid() && f.CanSet() {
			f.Set(reflect.ValueOf(classBase))
		}

		result := newElem.Interface()

		return result.(IClassBase), nil
	}

	return classBase, nil
}

type IClassType interface {
	Name() string
	Type() IType
	Properties() []IClassTypeProperty
	Constructors() []IClassTypeConstructor
	Methods() []IClassTypeMethod
	StaticMethods() []IClassTypeMethod
}

type IClassTypeConstructor interface {
	Name() string
	Symbol() string
	ArgumentTypes() []IType
}

type IClassTypeProperty interface {
	Name() string
	GetterType() IType
	SetterType() IType
	ReadOnly() bool
}

type IClassTypeMethod interface {
	Symbol() string
	ReturnType() IType
	ArgumentTypes() []IType
	IsOverload() bool
}

func (erc *classType) Name() string {
	return erc.legalFunctionName
}

func (erc *classType) Type() IType {
	return &exposedType{registeredType: erc}
}

func (erc *classType) Properties() []IClassTypeProperty {
	properties := make([]IClassTypeProperty, 0)

	for i := range erc.properties {
		properties = append(properties, erc.properties[i])
	}

	return properties
}

func (erc *classType) Methods() []IClassTypeMethod {
	methods := make([]IClassTypeMethod, 0)

	for i := range erc.methods {
		if erc.methods[i].isStatic {
			continue
		}

		if erc.methods[i].overloadTable != nil {
			for overload := range erc.methods[i].overloadTable {
				methods = append(methods, erc.methods[i].overloadTable[overload])
			}
		} else {
			methods = append(methods, erc.methods[i])
		}
	}

	return methods
}

func (erc *classType) StaticMethods() []IClassTypeMethod {
	methods := make([]IClassTypeMethod, 0)

	for i := range erc.methods {
		if !erc.methods[i].isStatic {
			continue
		}
		methods = append(methods, erc.methods[i])
	}

	return methods
}

type exposedClassConstructor struct {
	name             string
	classConstructor *classConstructor
}

func (ecc *exposedClassConstructor) Name() string {
	return ecc.name
}

func (ecc *exposedClassConstructor) Symbol() string {
	return ecc.name
}

func (ecc *exposedClassConstructor) ArgumentTypes() []IType {
	exposedTypes := make([]IType, len(ecc.classConstructor.argumentTypes))
	for i := range ecc.classConstructor.argumentTypes {
		exposedTypes[i] = &exposedType{ecc.classConstructor.argumentTypes[i]}
	}
	return exposedTypes
}

func (erc *classType) Constructors() []IClassTypeConstructor {
	constructors := make([]IClassTypeConstructor, 0)

	for i := range erc.constructors {
		constructor := &exposedClassConstructor{
			name:             "",
			classConstructor: erc.constructors[i],
		}

		if len(erc.constructors) > 1 {
			constructor.name = fmt.Sprintf("%d", i)
		}

		constructors = append(constructors, constructor)
	}

	return constructors
}

func (e *engine) GetClasses() []IClassType {
	classes := make([]IClassType, 0)
	for i := range e.registeredClasses {
		classes = append(classes, e.registeredClasses[i])
	}
	return classes
}

type ClassBase struct {
	engine                  *engine
	classType               *classType
	ptr                     uint32
	ptrType                 *registeredPointerType
	registeredPtrTypeRecord *registeredPointerTypeRecord
}

func (ecb *ClassBase) getClassType() *classType {
	return ecb.classType
}

func (ecb *ClassBase) String() string {
	return fmt.Sprintf("%s, ptr: %d", ecb.classType.name, ecb.ptr)
}

func (ecb *ClassBase) getPtr() uint32 {
	return ecb.ptr
}

func (ecb *ClassBase) getPtrType() *registeredPointerType {
	return ecb.ptrType
}

func (ecb *ClassBase) getRegisteredPtrTypeRecord() *registeredPointerTypeRecord {
	return ecb.registeredPtrTypeRecord
}

func (ecb *ClassBase) isValid() bool {
	return ecb != nil
}

func (ecb *ClassBase) Clone(ctx context.Context, this IClassBase) (IClassBase, error) {
	return ecb.classType.clone(ctx, this)
}

func (ecb *ClassBase) Delete(ctx context.Context, this IClassBase) error {
	return ecb.classType.delete(ctx, this)
}

func (ecb *ClassBase) CallMethod(ctx context.Context, this any, name string, arguments ...any) (any, error) {
	method, ok := ecb.classType.methods[name]
	if !ok {
		return nil, fmt.Errorf("method %s is not found on %T", name, this)
	}

	// Ensure that the engine is attached. Allows calling methods on the class
	// without keeping track of the engine.
	ctx = ecb.engine.Attach(ctx)
	return method.fn(ctx, this, arguments...)
}

func (ecb *ClassBase) SetProperty(ctx context.Context, this any, name string, value any) error {
	property, ok := ecb.classType.properties[name]
	if !ok {
		return fmt.Errorf("property %s is not found on %T", name, this)
	}

	// Ensure that the engine is attached. Allows setting properties on the
	// class without keeping track of the engine.
	ctx = ecb.engine.Attach(ctx)
	return property.set(ctx, this, value)
}

func (ecb *ClassBase) GetProperty(ctx context.Context, this any, name string) (any, error) {
	property, ok := ecb.classType.properties[name]
	if !ok {
		return nil, fmt.Errorf("property %s is not found on %T", name, this)
	}

	// Ensure that the engine is attached. Allows setting properties on the
	// class without keeping track of the engine.
	ctx = ecb.engine.Attach(ctx)
	return property.get(ctx, this)
}

type IClassBase interface {
	getClassType() *classType
	getPtr() uint32
	getPtrType() *registeredPointerType
	getRegisteredPtrTypeRecord() *registeredPointerTypeRecord
	isValid() bool
	Clone(ctx context.Context, this IClassBase) (IClassBase, error)
	Delete(ctx context.Context, this IClassBase) error
	CallMethod(ctx context.Context, this any, name string, arguments ...any) (any, error)
	SetProperty(ctx context.Context, this any, name string, value any) error
	GetProperty(ctx context.Context, this any, name string) (any, error)
}
