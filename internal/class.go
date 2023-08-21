package embind

import (
	"context"
	"fmt"
	"github.com/tetratelabs/wazero/api"
	"reflect"
)

type classProperty struct {
	enumerable   bool
	configurable bool
	readOnly     bool
	set          func(ctx context.Context, this any, v any) error
	get          func(ctx context.Context, this any) (any, error)
}

type classConstructor struct {
	fn         publicSymbolFn
	argTypes   []string
	resultType string
}

type classType struct {
	baseType
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
	return erc.name
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

func (erc *classType) isDeleted(handle IEmvalClassBase) bool {
	return handle.getRegisteredPtrTypeRecord().ptr == 0
}

func (erc *classType) deleteLater(handle IEmvalClassBase) (any, error) {
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

func (erc *classType) isAliasOf(ctx context.Context, first, second IEmvalClassBase) (bool, error) {
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

func (erc *classType) clone(from IEmvalClassBase) (IEmvalClassBase, error) {
	registeredPtrTypeRecord := from.getRegisteredPtrTypeRecord()
	if registeredPtrTypeRecord.ptr == 0 {
		return nil, fmt.Errorf("class handle already deleted")
	}

	if registeredPtrTypeRecord.preservePointerOnDelete {
		registeredPtrTypeRecord.count.value += 1
		return from, nil
	}

	clone, err := erc.getNewInstance(registeredPtrTypeRecord.shallowCopyInternalPointer())
	if err != nil {
		return nil, err
	}

	clone.getRegisteredPtrTypeRecord().count.value += 1
	clone.getRegisteredPtrTypeRecord().deleteScheduled = false
	return clone, nil
}

func (erc *classType) delete(ctx context.Context, handle IEmvalClassBase) error {
	registeredPtrTypeRecord := handle.getRegisteredPtrTypeRecord()
	if registeredPtrTypeRecord.ptr == 0 {
		return fmt.Errorf("class handle already deleted")
	}

	if registeredPtrTypeRecord.deleteScheduled && !registeredPtrTypeRecord.preservePointerOnDelete {
		return fmt.Errorf("object already scheduled for deletion")
	}

	err := registeredPtrTypeRecord.detachFinalizer(ctx)
	if err != nil {
		return err
	}

	err = registeredPtrTypeRecord.releaseClassHandle(ctx)
	if err != nil {
		return err
	}

	if registeredPtrTypeRecord.preservePointerOnDelete {
		registeredPtrTypeRecord.smartPtr = 0
		registeredPtrTypeRecord.ptr = 0
	}

	return nil
}

func (erc *classType) getNewInstance(record *registeredPointerTypeRecord) (IEmvalClassBase, error) {
	classBase := &EmvalClassBase{
		classType:               erc,
		ptr:                     record.ptr,
		ptrType:                 record.ptrType,
		registeredPtrTypeRecord: record,
	}

	// If we have a Go struct, wrap the resulting class in it.
	if erc.hasGoStruct {
		typeElem := reflect.TypeOf(erc.goStruct).Elem()
		newElem := reflect.New(typeElem)
		f := newElem.Elem().FieldByName("EmvalClassBase")
		if f.IsValid() && f.CanSet() {
			f.Set(reflect.ValueOf(classBase))
		}

		result := newElem.Interface()

		return result.(IEmvalClassBase), nil
	}

	return classBase, nil
}

type EmvalClassBase struct {
	engine                  *engine
	classType               *classType
	ptr                     uint32
	ptrType                 *registeredPointerType
	registeredPtrTypeRecord *registeredPointerTypeRecord
}

func (ecb *EmvalClassBase) getClassType() *classType {
	return ecb.classType
}

func (ecb *EmvalClassBase) getPtr() uint32 {
	return ecb.ptr
}

func (ecb *EmvalClassBase) getPtrType() *registeredPointerType {
	return ecb.ptrType
}

func (ecb *EmvalClassBase) getRegisteredPtrTypeRecord() *registeredPointerTypeRecord {
	return ecb.registeredPtrTypeRecord
}

func (ecb *EmvalClassBase) isValid() bool {
	return ecb != nil
}

func (ecb *EmvalClassBase) Clone(from IEmvalClassBase) (IEmvalClassBase, error) {
	return ecb.classType.clone(from)
}

func (ecb *EmvalClassBase) Delete(ctx context.Context, handle IEmvalClassBase) error {
	return ecb.classType.delete(ctx, handle)
}

func (ecb *EmvalClassBase) CallMethod(ctx context.Context, this any, name string, arguments ...any) (any, error) {
	method, ok := ecb.classType.methods[name]
	if !ok {
		return nil, fmt.Errorf("method %s is not found on %T", name, this)
	}

	// Ensure that the engine is attached. Allows calling methods on the class
	// without keeping track of the engine.
	ctx = ecb.engine.Attach(ctx)
	return method.fn(ctx, this, arguments)
}

func (ecb *EmvalClassBase) SetProperty(ctx context.Context, this any, name string, value any) error {
	property, ok := ecb.classType.properties[name]
	if !ok {
		return fmt.Errorf("property %s is not found on %T", name, this)
	}

	// Ensure that the engine is attached. Allows setting properties on the
	// class without keeping track of the engine.
	ctx = ecb.engine.Attach(ctx)
	return property.set(ctx, this, value)
}

func (ecb *EmvalClassBase) GetProperty(ctx context.Context, this any, name string) (any, error) {
	property, ok := ecb.classType.properties[name]
	if !ok {
		return nil, fmt.Errorf("property %s is not found on %T", name, this)
	}

	// Ensure that the engine is attached. Allows setting properties on the
	// class without keeping track of the engine.
	ctx = ecb.engine.Attach(ctx)
	return property.get(ctx, this)
}

type IEmvalClassBase interface {
	getClassType() *classType
	getPtr() uint32
	getPtrType() *registeredPointerType
	getRegisteredPtrTypeRecord() *registeredPointerTypeRecord
	isValid() bool
	Clone(from IEmvalClassBase) (IEmvalClassBase, error)
	Delete(ctx context.Context, handle IEmvalClassBase) error
	CallMethod(ctx context.Context, this any, name string, arguments ...any) (any, error)
	SetProperty(ctx context.Context, this any, name string, value any) error
	GetProperty(ctx context.Context, this any, name string) (any, error)
}
