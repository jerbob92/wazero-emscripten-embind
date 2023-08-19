package embind

import (
	"context"
	"fmt"
	"github.com/tetratelabs/wazero/api"
	"reflect"
	"unicode"
)

type EmvalConstructor interface {
	New(argTypes []string, args ...any) (any, error)
}

type EmvalFunctionMapper interface {
	MapFunction(name string, returnType string, argTypes []string) (string, error)
}

type emvalType struct {
	baseType
	engine *engine
}

func (et *emvalType) FromWireType(ctx context.Context, mod api.Module, value uint64) (any, error) {
	rv, err := et.engine.emvalEngine.toValue(api.DecodeI32(value))
	if err != nil {
		return nil, err
	}

	err = et.engine.emvalEngine.allocator.decref(api.DecodeI32(value))
	if err != nil {
		return nil, err
	}

	return rv, nil
}

func (et *emvalType) ToWireType(ctx context.Context, mod api.Module, destructors *[]*destructorFunc, o any) (uint64, error) {
	return api.EncodeI32(et.engine.emvalEngine.toHandle(o)), nil
}

func (et *emvalType) ReadValueFromPointer(ctx context.Context, mod api.Module, pointer uint32) (any, error) {
	value, ok := mod.Memory().ReadUint32Le(pointer)
	if !ok {
		return nil, fmt.Errorf("could not read emval value at pointer %d", pointer)
	}
	return et.FromWireType(ctx, mod, api.EncodeU32(value))
}

func (et *emvalType) GoType() string {
	return "int32"
}

type emvalHandle struct {
	value    any
	refCount int
}

type emvalAllocator struct {
	allocated []*emvalHandle
	freelist  []int32
	reserved  int
}

func (ea *emvalAllocator) get(id int32) (*emvalHandle, error) {
	if id < 1 || int(id) > len(ea.allocated)-1 {
		return nil, fmt.Errorf("invalid id: %d", id)
	}

	return ea.allocated[int(id)], nil
}

func (ea *emvalAllocator) has(id int32) bool {
	if id <= 1 || int(id) > ea.reserved-1 {
		return false
	}

	return true
}

func (ea *emvalAllocator) allocate(handle *emvalHandle) int32 {
	var id int32

	// Reuse items to free when available
	if len(ea.freelist) > 0 {
		// Get ID of last item.
		id = ea.freelist[len(ea.freelist)-1]

		// Remove the item that we just took.
		ea.freelist = ea.freelist[:len(ea.freelist)-1]

		ea.allocated[id] = handle
	} else {
		id = int32(len(ea.allocated))
		ea.allocated = append(ea.allocated, handle)
	}

	return id
}

func (ea *emvalAllocator) free(id int32) error {
	if id <= 1 || int(id) > len(ea.allocated)-1 {
		return fmt.Errorf("invalid id: %d", id)
	}

	// Set the slot to `undefined` rather than using `delete` here since
	// apparently arrays with holes in them can be less efficient.
	ea.allocated[id] = nil
	ea.freelist = append(ea.freelist, id)

	return nil
}

func (ea *emvalAllocator) incref(id int32) error {
	if id > 4 {
		handle, err := ea.get(id)
		if err != nil {
			return err
		}
		handle.refCount++
	}

	return nil
}

func (ea *emvalAllocator) decref(id int32) error {
	if int(id) >= ea.reserved {
		handle, err := ea.get(id)
		if err != nil {
			return err
		}

		handle.refCount--
		if handle.refCount == 0 {
			err = ea.free(id)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type emvalRegisteredMethod struct {
	id       int32
	argTypes []registeredType
	name     string
}

type emvalEngine struct {
	allocator             *emvalAllocator
	globals               map[string]any
	symbols               map[uint32]string
	registeredMethodCount int32
	registeredMethodIds   map[string]int32
	registeredMethods     map[int32]*emvalRegisteredMethod
}

func createEmvalEngine() *emvalEngine {
	return &emvalEngine{
		allocator: &emvalAllocator{
			allocated: []*emvalHandle{
				nil, // Reserve slot 0 so that 0 is always an invalid handle
				{
					value: undefined,
				},
				{
					value: nil,
				},
				{
					value: true,
				},
				{
					value: false,
				},
			},
			freelist: []int32{},
			reserved: 5,
		},
		globals:             map[string]any{},
		symbols:             map[uint32]string{},
		registeredMethodIds: map[string]int32{},
		registeredMethods:   map[int32]*emvalRegisteredMethod{},
	}
}

func (e *emvalEngine) toHandle(value any) int32 {
	if value == undefined {
		return 1
	} else if value == nil {
		return 2
	} else if value == true {
		return 3
	} else if value == true {
		return 4
	}
	return e.allocator.allocate(&emvalHandle{refCount: 1, value: value})
}

func (e *emvalEngine) toValue(id int32) (any, error) {
	handle, err := e.allocator.get(id)
	if err != nil {
		return nil, err
	}

	return handle.value, nil
}

func (e *emvalEngine) getGlobal(name string) any {
	global, ok := e.globals[name]
	if !ok {
		return undefined
	}
	return global
}

func (e *emvalEngine) getSymbolElem(symbol any) (*reflect.Value, error) {
	elem := reflect.ValueOf(symbol)
	if elem.Kind() != reflect.Ptr && elem.Kind() != reflect.Struct {
		return nil, fmt.Errorf("symbol is not a pointer or a struct, but a %s", reflect.TypeOf(symbol).Kind().String())
	}

	// Get elem behind pointer.
	if elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}

	if elem.Kind() != reflect.Struct {
		return nil, fmt.Errorf("symbol reference is not to a struct, but to a %s", elem.Kind().String())
	}

	return &elem, nil
}

func (e *emvalEngine) getElemField(symbol any, field string) (*reflect.Value, error) {
	elem, err := e.getSymbolElem(symbol)
	if err != nil {
		return nil, fmt.Errorf("no valid field %s on emval %T: %w", field, symbol, err)
	}

	typeElem := reflect.TypeOf(symbol).Elem()
	for i := 0; i < typeElem.NumField(); i++ {
		val := typeElem.Field(i)
		if val.Tag.Get("embind_property") == field {
			f := elem.FieldByName(val.Name)
			if f.IsValid() && f.CanSet() {
				return &f, nil
			}
		}
	}

	f := elem.FieldByName(field)
	if f.IsValid() && f.CanSet() {
		return &f, nil
	}

	upperFirst := string(unicode.ToUpper(rune(field[0]))) + field[1:]
	f = elem.FieldByName(upperFirst)
	if f.IsValid() && f.CanSet() {
		return &f, nil
	}

	return nil, fmt.Errorf("could not find field \"%s\" by embind_property tag, name or by %s", field, upperFirst)
}

func (e *emvalEngine) callMethod(ctx context.Context, mod api.Module, registeredMethod *emvalRegisteredMethod, handle any, methodName string, destructorsRef, argsBase uint32) (uint64, error) {
	var matchedMethod *reflect.Method
	st := reflect.TypeOf(handle)

	c, ok := handle.(EmvalFunctionMapper)
	if ok {
		argCount := len(registeredMethod.argTypes)
		argTypeNames := make([]string, argCount-1)
		for i := 1; i < argCount; i++ {
			argTypeNames[i-1] = registeredMethod.argTypes[i].Name()
		}
		mappedFunction, err := c.MapFunction(methodName, registeredMethod.argTypes[0].Name(), argTypeNames)
		if err != nil {
			return 0, fmt.Errorf("mapper function of type %T returned error: %w", handle, err)
		}

		if mappedFunction != "" {
			m, ok := st.MethodByName(mappedFunction)
			if ok {
				matchedMethod = &m
			} else {
				return 0, fmt.Errorf("mapper function of type %T returned method %s, but method could not be found", handle, mappedFunction)
			}
		}
	}

	actualMethodName := methodName
	if matchedMethod == nil {
		m, ok := st.MethodByName(methodName)
		if ok {
			matchedMethod = &m
		} else {
			actualMethodName = string(unicode.ToUpper(rune(methodName[0]))) + methodName[1:]
			m, ok = st.MethodByName(actualMethodName)
			if ok {
				matchedMethod = &m
			}
		}
	}

	if matchedMethod == nil {
		return 0, fmt.Errorf("type %T does not have method name %s or %s and was also not mapped by the EmvalFunctionMapper interface", handle, methodName, actualMethodName)
	}

	if !matchedMethod.IsExported() {
		return 0, fmt.Errorf("the method name %s on type %T is not exported", methodName, handle)
	}

	var err error
	argCount := len(registeredMethod.argTypes)
	args := make([]any, argCount)
	for i := 1; i < argCount; i++ {
		args[i], err = registeredMethod.argTypes[i].ReadValueFromPointer(ctx, mod, argsBase)
		if err != nil {
			return 0, fmt.Errorf("could not read arg value from pointer for arg %d, %w", i-1, err)
		}

		argsBase += uint32(registeredMethod.argTypes[i].ArgPackAdvance())
	}

	var destructors *[]*destructorFunc
	if destructorsRef != 0 {
		var destructors = &[]*destructorFunc{}
		rd := e.toHandle(destructors)
		ok = mod.Memory().WriteUint32Le(destructorsRef, uint32(rd))
		if !ok {
			return 0, fmt.Errorf("could not write destructor ref to memory")
		}
	}

	callArgs := make([]reflect.Value, argCount-1)

	// @todo: make sure that we got this right:
	/*
		// For a non-interface type T or *T, the returned Method's Type and Func
		// fields describe a function whose first argument is the receiver.
		//
		// For an interface type, the returned Method's Type field gives the
		// method signature, without a receiver, and the Func field is nil.
	*/
	//callArgs[0] = reflect.ValueOf(handle)

	for i := 1; i < argCount; i++ {
		callArgs[i-1] = reflect.ValueOf(args[i])
	}

	resultData := reflect.ValueOf(handle).MethodByName(matchedMethod.Name).Call(callArgs)

	for i := 1; i < argCount; i++ {
		if registeredMethod.argTypes[i].HasDeleteObject() {
			err = registeredMethod.argTypes[i].DeleteObject(ctx, mod, args[i])
			if err != nil {
				return 0, fmt.Errorf("could not delete object")
			}
		}
	}

	_, ok = registeredMethod.argTypes[0].(*voidType)
	if ok {
		return 0, nil
	}

	if len(resultData) == 0 {
		panic(fmt.Errorf("wrong result type count, got %d, need %d", len(resultData), 1))
	}

	rv := resultData[0].Interface()
	res, err := registeredMethod.argTypes[0].ToWireType(ctx, mod, destructors, rv)
	if err != nil {
		return 0, fmt.Errorf("could not call ToWireType on response")
	}

	return res, nil
}
