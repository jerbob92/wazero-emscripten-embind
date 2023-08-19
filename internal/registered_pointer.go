package embind

import (
	"context"
	"fmt"
	"github.com/tetratelabs/wazero/api"
)

type registeredPointerType struct {
	baseType
	registeredClass *classType
	isReference     bool
	isConst         bool

	// smart pointer properties
	isSmartPointer bool
	pointeeType    *registeredPointerType
	sharingPolicy  any
	rawGetPointee  api.Function
	rawConstructor api.Function
	rawShare       api.Function
	rawDestructor  api.Function
}

type registeredPointerTypeRecordCount struct {
	value int32
}

type registeredPointerTypeRecord struct {
	ptrType                 *registeredPointerType
	ptr                     uint32
	smartPtrType            *registeredPointerType
	smartPtr                uint32
	count                   *registeredPointerTypeRecordCount
	preservePointerOnDelete bool
	deleteScheduled         bool
}

func (rptr *registeredPointerTypeRecord) detachFinalizer(ctx context.Context) error {
	// @todo: what does this do?
	return nil
}

func (rptr *registeredPointerTypeRecord) shallowCopyInternalPointer() *registeredPointerTypeRecord {
	return &registeredPointerTypeRecord{
		count:                   rptr.count,
		deleteScheduled:         rptr.deleteScheduled,
		preservePointerOnDelete: rptr.preservePointerOnDelete,
		ptr:                     rptr.ptr,
		ptrType:                 rptr.ptrType,
		smartPtr:                rptr.smartPtr,
		smartPtrType:            rptr.smartPtrType,
	}
}

func (rptr *registeredPointerTypeRecord) releaseClassHandle(ctx context.Context) error {
	rptr.count.value -= 1
	toDelete := 0 == rptr.count.value
	if toDelete {
		return rptr.runDestructor(ctx)
	}
	return nil
}

func (rptr *registeredPointerTypeRecord) runDestructor(ctx context.Context) error {
	if rptr.smartPtr != 0 {
		_, err := rptr.smartPtrType.rawDestructor.Call(ctx, api.EncodeU32(rptr.smartPtr))
		if err != nil {
			return err
		}
	} else {
		_, err := rptr.ptrType.registeredClass.rawDestructor.Call(ctx, api.EncodeU32(rptr.ptr))
		if err != nil {
			return err
		}
	}
	return nil
}

func (rpt *registeredPointerType) FromWireType(ctx context.Context, mod api.Module, value uint64) (any, error) {
	// ptr is a raw pointer (or a raw smartpointer)
	ptr := api.DecodeU32(value)

	// rawPointer is a maybe-null raw pointer
	rawPointer, err := rpt.getPointee(ctx, ptr)
	if err != nil {
		return nil, err
	}

	if rawPointer == 0 {
		destrFun, err := rpt.DestructorFunction(ctx, mod, ptr)
		if err != nil {
			return nil, err
		}
		if destrFun != nil {
			err = destrFun.run(ctx, mod)
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	}

	registeredInstance, err := rpt.getInheritedInstance(ctx, mod, rpt.registeredClass, rawPointer)
	if err != nil {
		return nil, err
	}

	if registeredInstance != nil {
		if registeredInstance.RegisteredPtrTypeRecord().count.value == 0 {
			registeredInstance.RegisteredPtrTypeRecord().ptr = rawPointer
			registeredInstance.RegisteredPtrTypeRecord().smartPtr = ptr
			return registeredInstance.ClassType().clone(registeredInstance)
		} else {
			// else, just increment reference count on existing object
			// it already has a reference to the smart pointer
			rv, err := registeredInstance.ClassType().clone(registeredInstance)
			if err != nil {
				return nil, err
			}
			destructor, err := rpt.DestructorFunction(ctx, mod, ptr)
			if err != nil {
				return nil, err
			}
			err = destructor.run(ctx, mod)
			if err != nil {
				return nil, err
			}

			return rv, nil
		}
	}

	makeDefaultHandle := func() (any, error) {
		if rpt.isSmartPointer {
			return rpt.makeClassHandle(rpt.registeredClass, &registeredPointerTypeRecord{
				ptrType:      rpt.pointeeType,
				ptr:          rawPointer,
				smartPtrType: rpt,
				smartPtr:     ptr,
			})
		} else {
			return rpt.makeClassHandle(rpt.registeredClass, &registeredPointerTypeRecord{
				ptrType: rpt,
				ptr:     ptr,
			})
		}
	}

	res, err := rpt.registeredClass.getActualType.Call(ctx, api.EncodeU32(rawPointer))
	if err != nil {
		return nil, err
	}
	actualType := api.DecodeI32(res[0])

	e := MustGetEngineFromContext(ctx, mod).(*engine)

	registeredPointerRecord, ok := e.registeredPointers[actualType]
	if !ok {
		defaultHandle, err := makeDefaultHandle()
		if err != nil {
			return nil, err
		}

		return defaultHandle, nil
	}

	var toType *registeredPointerType
	if rpt.isConst {
		toType = registeredPointerRecord.constPointerType
	} else {
		toType = registeredPointerRecord.pointerType
	}

	dp, err := rpt.downcastPointer(ctx, rawPointer, rpt.registeredClass, toType.registeredClass)
	if err != nil {
		return nil, err
	}

	if dp == 0 {
		defaultHandle, err := makeDefaultHandle()
		if err != nil {
			return nil, err
		}

		return defaultHandle, nil
	}

	if rpt.isSmartPointer {
		return rpt.makeClassHandle(toType.registeredClass, &registeredPointerTypeRecord{
			ptrType:      toType,
			ptr:          dp,
			smartPtrType: rpt,
			smartPtr:     ptr,
		})
	} else {
		return rpt.makeClassHandle(toType.registeredClass, &registeredPointerTypeRecord{
			ptrType: toType,
			ptr:     dp,
		})
	}
}

func (rpt *registeredPointerType) ToWireType(ctx context.Context, mod api.Module, destructors *[]*destructorFunc, o any) (uint64, error) {
	return 0, fmt.Errorf("unknown registered pointer toWireType")
}

func (rpt *registeredPointerType) ReadValueFromPointer(ctx context.Context, mod api.Module, pointer uint32) (any, error) {
	value, ok := mod.Memory().ReadUint32Le(pointer)
	if !ok {
		return nil, fmt.Errorf("could not read register pointer value at pointer %d", pointer)
	}
	return rpt.FromWireType(ctx, mod, api.EncodeU32(value))
}

func (rpt *registeredPointerType) HasDestructorFunction() bool {
	return true
}

func (rpt *registeredPointerType) DestructorFunction(ctx context.Context, mod api.Module, pointer uint32) (*destructorFunc, error) {
	if rpt.rawDestructor != nil {
		return &destructorFunc{
			apiFunction: rpt.rawDestructor,
			args:        []uint64{api.EncodeU32(pointer)},
		}, nil
	}

	return nil, nil
}

func (rpt *registeredPointerType) HasDeleteObject() bool {
	return true
}

func (rpt *registeredPointerType) DeleteObject(ctx context.Context, mod api.Module, handle any) error {
	if handle != nil {
		casted := handle.(IEmvalClassBase)
		return casted.ClassType().delete(ctx, casted)
	}
	return nil
}

func (rpt *registeredPointerType) getPointee(ctx context.Context, ptr uint32) (uint32, error) {
	if rpt.rawGetPointee != nil {
		res, err := rpt.rawGetPointee.Call(ctx, api.EncodeU32(ptr))
		if err != nil {
			return 0, err
		}
		ptr = api.DecodeU32(res[0])
	}
	return ptr, nil
}

func (rpt *registeredPointerType) getBasestPointer(ctx context.Context, class *classType, ptr uint32) (uint32, error) {
	if ptr == 0 {
		return 0, fmt.Errorf("ptr should not be 0")
	}
	for class.baseClass != nil {
		res, err := class.upcast.Call(ctx, api.EncodeU32(ptr))
		if err != nil {
			return 0, nil
		}
		ptr = api.DecodeU32(res[0])
		class = class.baseClass
	}

	return ptr, nil
}

func (rpt *registeredPointerType) getInheritedInstance(ctx context.Context, mod api.Module, class *classType, ptr uint32) (IEmvalClassBase, error) {
	ptr, err := rpt.getBasestPointer(ctx, class, ptr)
	if err != nil {
		return nil, err
	}

	e := MustGetEngineFromContext(ctx, mod).(*engine)
	instance, ok := e.registeredInstances[ptr]
	if !ok {
		return nil, nil
	}

	return instance, nil
}

func (rpt *registeredPointerType) downcastPointer(ctx context.Context, ptr uint32, ptrClass *classType, desiredClass *classType) (uint32, error) {
	if ptrClass == desiredClass {
		return ptr, nil
	}

	if desiredClass.baseClass == nil {
		return 0, nil
	}

	rv, err := rpt.downcastPointer(ctx, ptr, ptrClass, desiredClass.baseClass)
	if err != nil {
		return 0, err
	}
	if rv == 0 {
		return 0, nil
	}

	downcastRes, err := desiredClass.downcast.Call(ctx, api.EncodeU32(rv))
	if err != nil {
		return 0, err
	}

	return api.DecodeU32(downcastRes[0]), nil
}

func (rpt *registeredPointerType) makeClassHandle(class *classType, record *registeredPointerTypeRecord) (IEmvalClassBase, error) {
	if record.ptrType == nil || record.ptr == 0 {
		return nil, fmt.Errorf("makeClassHandle requires ptr and ptrType")
	}
	hasSmartPtrType := record.smartPtrType != nil
	hasSmartPtr := record.smartPtr != 0
	if hasSmartPtrType != hasSmartPtr {
		return nil, fmt.Errorf("both smartPtrType and smartPtr must be specified")
	}
	record.count = &registeredPointerTypeRecordCount{
		value: 1,
	}

	if !class.hasGoStruct {
		return nil, fmt.Errorf("no struct registered for class %s", class.name)
	}

	classHandle, err := class.getInstanceFromGoStruct(record)
	if err != nil {
		return nil, err
	}

	rpt.attachFinalizer(classHandle)

	return classHandle, nil
}

func (rpt *registeredPointerType) attachFinalizer(classHandle any) {
	// @todo: attach Go GC for garbage collection?
	/**
	  if ('undefined' === typeof FinalizationRegistry) {
	    attachFinalizer = (handle) => handle;
	    return handle;
	  }
	  // If the running environment has a FinalizationRegistry (see
	  // https://github.com/tc39/proposal-weakrefs), then attach finalizers
	  // for class handles.  We check for the presence of FinalizationRegistry
	  // at run-time, not build-time.
	  finalizationRegistry = new FinalizationRegistry((info) => {
	    console.warn(info.leakWarning.stack.replace(/^Error: /, ''));
	    releaseClassHandle(info.$$);
	  });
	  attachFinalizer = (handle) => {
	    var $$ = handle.$$;
	    var hasSmartPtr = !!$$.smartPtr;
	    if (hasSmartPtr) {
	      // We should not call the destructor on raw pointers in case other code expects the pointee to live
	      var info = { $$: $$ };
	      // Create a warning as an Error instance in advance so that we can store
	      // the current stacktrace and point to it when / if a leak is detected.
	      // This is more useful than the empty stacktrace of `FinalizationRegistry`
	      // callback.
	      var cls = $$.ptrType.registeredClass;
	      info.leakWarning = new Error(`Embind found a leaked C++ instance ${cls.name} <${ptrToString($$.ptr)}>.\n` +
	      "We'll free it automatically in this case, but this functionality is not reliable across various environments.\n" +
	      "Make sure to invoke .delete() manually once you're done with the instance instead.\n" +
	      "Originally allocated"); // `.stack` will add "at ..." after this sentence
	      if ('captureStackTrace' in Error) {
	        Error.captureStackTrace(info.leakWarning, RegisteredPointer_fromWireType);
	      }
	      finalizationRegistry.register(handle, info, handle);
	    }
	    return handle;
	  };
	  detachFinalizer = (handle) => finalizationRegistry.unregister(handle);
	  return attachFinalizer(handle);
	*/
}
