package embind

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/jerbob92/wazero-emscripten-embind/types"

	"github.com/tetratelabs/wazero/api"
)

type IEmvalConstructor interface {
	New(argTypes []string, args ...any) (any, error)
}

type IEmvalFunctionMapper interface {
	MapFunction(name string, returnType string, argTypes []string) (string, error)
}

type emvalType struct {
	baseType
}

func (et *emvalType) FromWireType(ctx context.Context, mod api.Module, value uint64) (any, error) {
	e := MustGetEngineFromContext(ctx, mod).(*engine)
	rv, err := e.emvalEngine.toValue(api.DecodeI32(value))
	if err != nil {
		return nil, err
	}

	err = e.emvalEngine.allocator.decref(api.DecodeI32(value))
	if err != nil {
		return nil, err
	}

	return rv, nil
}

func (et *emvalType) ToWireType(ctx context.Context, mod api.Module, destructors *[]*destructorFunc, o any) (uint64, error) {
	e := MustGetEngineFromContext(ctx, mod).(*engine)
	return api.EncodeI32(e.emvalEngine.toHandle(o)), nil
}

func (et *emvalType) ReadValueFromPointer(ctx context.Context, mod api.Module, pointer uint32) (any, error) {
	value, ok := mod.Memory().ReadUint32Le(pointer)
	if !ok {
		return nil, fmt.Errorf("could not read emval value at pointer %d", pointer)
	}
	return et.FromWireType(ctx, mod, api.EncodeU32(value))
}

func (et *emvalType) GoType() string {
	return "any"
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
					value: types.Undefined,
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
	if value == types.Undefined {
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
		return types.Undefined
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

	c, ok := handle.(IEmvalFunctionMapper)
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
		if len(resultData) > 1 {
			return 0, fmt.Errorf("wrong result type count, got %d, need at most 1 (error)", len(resultData))
		}

		if len(resultData) == 1 {
			if resultData[0].Interface() != nil {
				err, isError := resultData[0].Interface().(error)
				if isError {
					return 0, fmt.Errorf("function returned error: %w", err)
				} else {
					return 0, fmt.Errorf("function returned non-error value in error return: %v", resultData[0].Interface())
				}
			}
		}

		return 0, nil
	}

	if len(resultData) < 1 {
		return 0, fmt.Errorf("wrong result type count, got %d, need at least 1 and at most 2 (value and error)", len(resultData))
	}

	if len(resultData) == 2 {
		if resultData[1].Interface() != nil {
			err, isError := resultData[1].Interface().(error)
			if isError {
				return 0, fmt.Errorf("function returned error: %w", err)
			} else {
				return 0, fmt.Errorf("function returned non-error value in error return: %v", resultData[1].Interface())
			}
		}
	}

	rv := resultData[0].Interface()
	res, err := registeredMethod.argTypes[0].ToWireType(ctx, mod, destructors, rv)
	if err != nil {
		return 0, fmt.Errorf("could not call ToWireType on response")
	}

	return res, nil
}

var RegisterEmval = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)

	rawType := api.DecodeI32(stack[0])
	name, err := engine.readCString(uint32(api.DecodeI32(stack[1])))
	if err != nil {
		panic(fmt.Errorf("could not read name: %w", err))
	}

	err = engine.registerType(rawType, &emvalType{
		baseType: baseType{
			rawType:        rawType,
			name:           name,
			argPackAdvance: 8,
		},
	}, &registerTypeOptions{
		ignoreDuplicateRegistrations: true,
	})
	if err != nil {
		panic(fmt.Errorf("could not register: %w", err))
	}
})

var EmvalTakeValue = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)
	rawType := api.DecodeI32(stack[0])

	registeredType, ok := engine.registeredTypes[rawType]
	if !ok {
		typeName, err := engine.getTypeName(ctx, rawType)
		if err != nil {
			panic(err)
		}
		panic(fmt.Errorf("_emval_take_value has unknown type %s", typeName))
	}

	arg := api.DecodeI32(stack[1])
	value, err := registeredType.ReadValueFromPointer(ctx, mod, uint32(arg))
	if err != nil {
		panic(fmt.Errorf("could not take value for _emval_take_value: %w", err))
	}

	id := engine.emvalEngine.toHandle(value)
	stack[0] = api.EncodeI32(id)
})

var EmvalIncref = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)
	handle := api.DecodeI32(stack[0])
	err := engine.emvalEngine.allocator.incref(handle)
	if err != nil {
		panic(fmt.Errorf("could not emval incref: %w", err))
	}
})

var EmvalDecref = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)
	handle := api.DecodeI32(stack[0])
	err := engine.emvalEngine.allocator.decref(handle)
	if err != nil {
		panic(fmt.Errorf("could not emval incref: %w", err))
	}
})

var EmvalRegisterSymbol = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)
	address := uint32(api.DecodeI32(stack[0]))
	name, err := engine.readCString(address)
	if err != nil {
		panic(fmt.Errorf("could not get symbol name"))
	}
	engine.emvalEngine.symbols[address] = name
})

var EmvalGetGlobal = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)
	name := api.DecodeI32(stack[0])

	if name == 0 {
		stack[0] = api.EncodeI32(engine.emvalEngine.toHandle(engine.emvalEngine.getGlobal("")))
	} else {
		name, err := engine.getStringOrSymbol(uint32(name))
		if err != nil {
			panic(fmt.Errorf("could not get symbol name"))
		}
		stack[0] = api.EncodeI32(engine.emvalEngine.toHandle(engine.emvalEngine.getGlobal(name)))
	}
})

var EmvalAs = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)
	id := api.DecodeI32(stack[0])
	handle, err := engine.emvalEngine.toValue(id)
	if err != nil {
		panic(fmt.Errorf("could not find handle: %w", err))
	}

	returnType, err := engine.requireRegisteredType(ctx, api.DecodeI32(stack[1]), "emval::as")
	if err != nil {
		panic(fmt.Errorf("could not require registered type: %w", err))
	}

	var destructors = &[]*destructorFunc{}
	rd := engine.emvalEngine.toHandle(destructors)
	ok := mod.Memory().WriteUint32Le(uint32(api.DecodeI32(stack[2])), uint32(rd))
	if !ok {
		panic(fmt.Errorf("could not write destructor ref to memory"))
	}

	returnVal, err := returnType.ToWireType(ctx, mod, destructors, handle)
	if err != nil {
		panic(fmt.Errorf("could not call toWireType on _emval_as: %w", err))
	}

	stack[0] = api.EncodeF64(returnType.ToF64(returnVal))
})

var EmvalNew = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)
	id := api.DecodeI32(stack[0])

	handle, err := engine.emvalEngine.toValue(id)
	if err != nil {
		panic(fmt.Errorf("could not get value of handle: %w", err))
	}

	argCount := int(api.DecodeI32(stack[1]))
	argsTypeBase := uint32(api.DecodeI32(stack[2]))
	argsBase := uint32(api.DecodeI32(stack[3]))

	args := make([]any, argCount)
	argTypeNames := make([]string, argCount)
	for i := 0; i < argCount; i++ {
		argType, ok := mod.Memory().ReadUint32Le(argsTypeBase + (4 * uint32(i)))
		if !ok {
			panic(fmt.Errorf("could not read arg type for arg %d from memory", i))
		}

		registeredArgType, err := engine.requireRegisteredType(ctx, int32(argType), fmt.Sprintf("argument %d", i))
		if err != nil {
			panic(fmt.Errorf("could not require registered type: %w", err))
		}

		args[i], err = registeredArgType.ReadValueFromPointer(ctx, mod, argsBase)
		if err != nil {
			panic(fmt.Errorf("could not read arg value from pointer for arg %d, %w", i, err))
		}

		argsBase += uint32(registeredArgType.ArgPackAdvance())

		argTypeNames[i] = registeredArgType.Name()
	}

	var res any
	c, ok := handle.(IEmvalConstructor)
	if ok {
		res, err = c.New(argTypeNames, args...)
		if err != nil {
			panic(fmt.Errorf("could not instaniate new value on %T with New(): %w", handle, err))
		}
	} else {
		typeElem := reflect.TypeOf(handle)

		// If we received a pointer, resolve the struct behind it.
		if typeElem.Kind() == reflect.Pointer {
			typeElem = typeElem.Elem()
		}

		// Make new instance of struct.
		newElem := reflect.New(typeElem)

		// Set the values on the struct if we need to/can.
		if argCount > 0 {
			if typeElem.Kind() != reflect.Struct {
				panic(fmt.Errorf("could not instaniate new value of %T: arguments required but can only be set on a struct", handle))
			}

			for i := 0; i < argCount; i++ {
				argSet := false
				for fieldI := 0; fieldI < typeElem.NumField(); fieldI++ {
					var err error
					func() {
						defer func() {
							if recoverErr := recover(); recoverErr != nil {
								realError, ok := recoverErr.(error)
								if ok {
									err = fmt.Errorf("could not set arg %d with embind_arg tag on emval %T: %w", i, handle, realError)
								}
								err = fmt.Errorf("could not set arg %d with embind_arg tag on emval %T: %v", i, handle, recoverErr)
							}
						}()

						val := typeElem.Field(fieldI)
						if val.Tag.Get("embind_arg") == strconv.Itoa(i) {
							f := newElem.Elem().FieldByName(val.Name)
							if f.IsValid() && f.CanSet() {
								f.Set(reflect.ValueOf(args[i]))
								argSet = true
							}
						}
					}()
					if err != nil {
						panic(fmt.Errorf("could not instaniate new value of %T: %w", handle, err))
					}
				}
				if !argSet {
					panic(fmt.Errorf("could not instaniate new value of %T: could not bind arg %d", handle, i))
				}
			}
		}

		if reflect.TypeOf(handle).Kind() == reflect.Pointer {
			res = newElem.Interface()
		} else {
			res = newElem.Elem().Interface()
		}
	}

	stack[0] = api.EncodeI32(engine.emvalEngine.toHandle(res))
})

var EmvalSetProperty = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)

	handle, err := engine.emvalEngine.toValue(api.DecodeI32(stack[0]))
	if err != nil {
		panic(fmt.Errorf("could not find handle: %w", err))
	}

	key, err := engine.emvalEngine.toValue(api.DecodeI32(stack[1]))
	if err != nil {
		panic(fmt.Errorf("could not find key: %w", err))
	}

	val, err := engine.emvalEngine.toValue(api.DecodeI32(stack[2]))
	if err != nil {
		panic(fmt.Errorf("could not find val: %w", err))
	}

	keyString, ok := key.(string)
	if !ok {
		panic(fmt.Errorf("could not set property on emval %T: %w", handle, errors.New("key is not of type string")))
	}

	f, err := engine.emvalEngine.getElemField(handle, keyString)
	if err != nil {
		panic(fmt.Errorf("could not set property %s on emval %T: %w", keyString, handle, err))
	}

	defer func() {
		if err := recover(); err != nil {
			realError, ok := err.(error)
			if ok {
				panic(fmt.Errorf("could not set property %s on emval %T: %w", keyString, handle, realError))
			}
			panic(fmt.Errorf("could not set property %s on emval %T: %v", keyString, handle, err))
		}
	}()

	f.Set(reflect.ValueOf(val))
})

var EmvalGetProperty = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)

	handle, err := engine.emvalEngine.toValue(api.DecodeI32(stack[0]))
	if err != nil {
		panic(fmt.Errorf("could not find handle: %w", err))
	}

	key, err := engine.emvalEngine.toValue(api.DecodeI32(stack[1]))
	if err != nil {
		panic(fmt.Errorf("could not find key: %w", err))
	}

	keyString, ok := key.(string)
	if !ok {
		panic(fmt.Errorf("could not get property on emval %T: %w", handle, errors.New("key is not of type string")))
	}

	f, err := engine.emvalEngine.getElemField(handle, keyString)
	if err != nil {
		panic(fmt.Errorf("could not get property %s on emval %T: %w", keyString, handle, err))
	}

	defer func() {
		if err := recover(); err != nil {
			realError, ok := err.(error)
			if ok {
				panic(fmt.Errorf("could not get property %s on emval %T: %w", keyString, handle, realError))
			}
			panic(fmt.Errorf("could not get property %s on emval %T: %v", keyString, handle, err))
		}
	}()

	stack[0] = api.EncodeI32(engine.emvalEngine.toHandle(f.Interface()))
})

var EmvalNewCString = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)
	v := api.DecodeI32(stack[0])
	name, err := engine.getStringOrSymbol(uint32(v))
	if err != nil {
		panic(fmt.Errorf("could not get symbol name"))
	}
	stack[0] = api.EncodeI32(engine.emvalEngine.toHandle(name))
})

var EmvalRunDestructors = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)
	id := api.DecodeI32(stack[0])
	destructorsVal, err := engine.emvalEngine.toValue(id)
	if err != nil {
		panic(fmt.Errorf("could not find handle: %w", err))
	}

	destructors := destructorsVal.(*[]*destructorFunc)

	err = engine.runDestructors(ctx, *destructors)
	if err != nil {
		panic(fmt.Errorf("could not run destructors: %w", err))
	}

	err = engine.emvalEngine.allocator.decref(id)
	if err != nil {
		panic(fmt.Errorf("could not run decref id %d: %w", id, err))
	}
})

var EmvalGetMethodCaller = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)

	argCount := int(api.DecodeI32(stack[0]))
	argsTypeBase := uint32(api.DecodeI32(stack[1]))

	typeNames := make([]string, argCount)
	argTypes := make([]registeredType, argCount)
	for i := 0; i < argCount; i++ {
		argType, ok := mod.Memory().ReadUint32Le(argsTypeBase + (4 * uint32(i)))
		if !ok {
			panic(fmt.Errorf("could not read arg type for arg %d from memory", i))
		}

		registeredType, err := engine.requireRegisteredType(ctx, int32(argType), fmt.Sprintf("argument %d", i))
		if err != nil {
			panic(fmt.Errorf("could not require registered type: %w", err))
		}

		typeNames[i] = registeredType.Name()
		argTypes[i] = registeredType
	}

	signatureName := typeNames[0] + "_$" + strings.Join(typeNames[1:], "_") + "$"

	id, ok := engine.emvalEngine.registeredMethodIds[signatureName]
	if ok {
		stack[0] = api.EncodeI32(id)
		return
	}

	newID := engine.emvalEngine.registeredMethodCount
	newRegisteredMethod := &emvalRegisteredMethod{
		id:       newID,
		argTypes: argTypes,
		name:     signatureName,
	}
	engine.emvalEngine.registeredMethodIds[signatureName] = newID
	engine.emvalEngine.registeredMethods[newID] = newRegisteredMethod
	engine.emvalEngine.registeredMethodCount++

	stack[0] = api.EncodeI32(newID)
	return
})

var EmvalCall = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)
	id := api.DecodeI32(stack[0])
	argCount := api.DecodeI32(stack[1])
	argTypes := api.DecodeI32(stack[2])
	argv := api.DecodeI32(stack[3])

	handle, err := engine.emvalEngine.toValue(id)
	if err != nil {
		panic(fmt.Errorf("could not find handle: %w", err))
	}

	registeredArgTypes, err := engine.lookupTypes(ctx, argCount, argTypes)
	if err != nil {
		panic(fmt.Errorf("could not load required types: %w", err))
	}

	args := make([]any, argCount)
	for i := 0; i < int(argCount); i++ {
		requiredType := registeredArgTypes[i]
		args[i], err = requiredType.ReadValueFromPointer(ctx, mod, uint32(argv))
		if err != nil {
			panic(fmt.Errorf("could not load argument value: %w", err))
		}

		argv += requiredType.ArgPackAdvance()
	}

	reflectValues := make([]reflect.Value, argCount)
	for i := range args {
		reflectValues[i] = reflect.ValueOf(args[i])
	}

	value := reflect.ValueOf(handle)
	result := value.Call(reflectValues)

	var resultVal any = types.Undefined
	if len(result) > 0 {
		resultVal = result[0].Interface()
	}

	newHandle := engine.emvalEngine.toHandle(resultVal)
	stack[0] = api.EncodeI32(newHandle)
})

var EmvalCallMethod = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)
	caller := api.DecodeI32(stack[0])

	registeredMethod, ok := engine.emvalEngine.registeredMethods[caller]
	if !ok {
		panic(fmt.Errorf("could not call method with ID %d", caller))
	}

	id := api.DecodeI32(stack[1])
	handle, err := engine.emvalEngine.toValue(id)
	if err != nil {
		panic(fmt.Errorf("could not find handle: %w", err))
	}

	methodName, err := engine.getStringOrSymbol(uint32(api.DecodeI32(stack[2])))
	if err != nil {
		panic(fmt.Errorf("could not get symbol name"))
	}

	argsBase := uint32(api.DecodeI32(stack[4]))
	destructorsRef := uint32(api.DecodeI32(stack[3]))

	res, err := engine.emvalEngine.callMethod(ctx, mod, registeredMethod, handle, methodName, destructorsRef, argsBase)
	if err != nil {
		panic(fmt.Errorf("could not call %s on %T: %w", methodName, handle, err))
	}
	stack[0] = api.EncodeF64(float64(res))
})

var EmvalCallVoidMethod = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)
	caller := api.DecodeI32(stack[0])

	registeredMethod, ok := engine.emvalEngine.registeredMethods[caller]
	if !ok {
		panic(fmt.Errorf("could not call method with ID %d", caller))
	}

	id := api.DecodeI32(stack[1])
	handle, err := engine.emvalEngine.toValue(id)
	if err != nil {
		panic(fmt.Errorf("could not find handle: %w", err))
	}

	methodName, err := engine.getStringOrSymbol(uint32(api.DecodeI32(stack[2])))
	if err != nil {
		panic(fmt.Errorf("could not get symbol name"))
	}

	argsBase := uint32(api.DecodeI32(stack[3]))

	_, err = engine.emvalEngine.callMethod(ctx, mod, registeredMethod, handle, methodName, 0, argsBase)
	if err != nil {
		panic(fmt.Errorf("could not call %s on %T: %w", methodName, handle, err))
	}
})

var EmvalInstanceof = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalInstanceof call unimplemented")
})

var EmvalTypeof = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)
	id := api.DecodeI32(stack[0])
	handle, err := engine.emvalEngine.toValue(id)
	if err != nil {
		panic(fmt.Errorf("could not find handle: %w", err))
	}

	// Default type.
	typeOf := "object"

	if handle != nil {
		reflectTypeOf := reflect.TypeOf(handle)
		switch reflectTypeOf.Kind() {
		case reflect.Func:
			typeOf = "function"
		case reflect.String:
			typeOf = "string"
		case reflect.Bool:
			typeOf = "boolean"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
			reflect.Uintptr, reflect.Float32, reflect.Float64:
			typeOf = "number"
		case reflect.Int64, reflect.Uint64:
			typeOf = "bigint"
		}

		if handle == types.Undefined {
			typeOf = "undefined"
		}
	}

	stack[0] = api.EncodeI32(engine.emvalEngine.toHandle(typeOf))
})

var EmvalAsInt64 = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)
	id := api.DecodeI32(stack[0])
	handle, err := engine.emvalEngine.toValue(id)
	if err != nil {
		panic(fmt.Errorf("could not find handle: %w", err))
	}

	returnType, err := engine.requireRegisteredType(ctx, api.DecodeI32(stack[1]), "emval::as")
	if err != nil {
		panic(fmt.Errorf("could not require registered type: %w", err))
	}

	returnVal, err := returnType.ToWireType(ctx, mod, nil, handle)
	if err != nil {
		panic(fmt.Errorf("could not call toWireType on _emval_as: %w", err))
	}

	stack[0] = returnVal
})

var EmvalAsUint64 = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	engine := MustGetEngineFromContext(ctx, mod).(*engine)
	id := api.DecodeI32(stack[0])
	handle, err := engine.emvalEngine.toValue(id)
	if err != nil {
		panic(fmt.Errorf("could not find handle: %w", err))
	}

	returnType, err := engine.requireRegisteredType(ctx, api.DecodeI32(stack[1]), "emval::as")
	if err != nil {
		panic(fmt.Errorf("could not require registered type: %w", err))
	}

	returnVal, err := returnType.ToWireType(ctx, mod, nil, handle)
	if err != nil {
		panic(fmt.Errorf("could not call toWireType on _emval_as: %w", err))
	}

	stack[0] = returnVal
})

var EmvalAwait = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalAwait call unimplemented")
})

var EmvalDelete = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalDelete call unimplemented")
})

var EmvalEquals = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalEquals call unimplemented")
})

var EmvalGetModuleProperty = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalGetModuleProperty call unimplemented")
})

var EmvalIn = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalIn call unimplemented")
})

var EmvalIsNumber = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalIsNumber call unimplemented")
})

var EmvalIsString = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalIsString call unimplemented")
})

var EmvalLessThan = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalLessThan call unimplemented")
})

var EmvalNewArray = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	e := MustGetEngineFromContext(ctx, mod).(*engine)
	stack[0] = api.EncodeI32(e.emvalEngine.toHandle([]any{}))
})

var EmvalNewArrayFromMemoryView = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalNewArrayFromMemoryView call unimplemented")
})

var EmvalNewObject = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalNewObject call unimplemented")
})

var EmvalNewU16string = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalNewU16string call unimplemented")
})

var EmvalNewU8string = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalNewU8string call unimplemented")
})

var EmvalNot = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalNot call unimplemented")
})

var EmvalStrictlyEquals = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalStrictlyEquals call unimplemented")
})

var EmvalThrow = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalThrow call unimplemented")
})

var EmvalGreaterThan = api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
	// @todo: implement me.
	panic("EmvalGreaterThan call unimplemented")
})
