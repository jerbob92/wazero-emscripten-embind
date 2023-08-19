package embind

import (
	"context"
	"fmt"
	"github.com/tetratelabs/wazero/api"
)

type classProperty struct {
	enumerable   bool
	configurable bool
	set          func(ctx context.Context, mod api.Module, this any, v any) error
	get          func(ctx context.Context, mod api.Module, this any) (any, error)
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
	pureVirtualFunctions []string
	methods              map[string]*publicSymbol
	properties           map[string]*classProperty
	constructors         map[int32]publicSymbolFn
}

func (erc *classType) FromWireType(ctx context.Context, mod api.Module, value uint64) (any, error) {
	panic("should not be called")
}

func (erc *classType) ToWireType(ctx context.Context, mod api.Module, destructors *[]*destructorFunc, o any) (uint64, error) {
	panic("should not be called")
}

func (erc *classType) ReadValueFromPointer(ctx context.Context, mod api.Module, pointer uint32) (any, error) {
	panic("should not be called")
}

func (erc *classType) validate() error {
	// @todo: implement validator here.
	// @todo: we want to check if the Go struct implements everything we need.
	return nil
}

func (erc *classType) isDeleted(handle IEmvalClassBase) bool {
	return handle.RegisteredPtrTypeRecord().ptr == 0
}

func (erc *classType) deleteLater(handle IEmvalClassBase) (any, error) {
	registeredPtrTypeRecord := handle.RegisteredPtrTypeRecord()
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
	leftClass := first.RegisteredPtrTypeRecord().ptrType.registeredClass
	left := first.RegisteredPtrTypeRecord().ptr
	rightClass := second.RegisteredPtrTypeRecord().ptrType.registeredClass
	right := second.RegisteredPtrTypeRecord().ptr

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
	registeredPtrTypeRecord := from.RegisteredPtrTypeRecord()
	if registeredPtrTypeRecord.ptr == 0 {
		return nil, fmt.Errorf("class handle already deleted")
	}

	if registeredPtrTypeRecord.preservePointerOnDelete {
		registeredPtrTypeRecord.count.value += 1
		return from, nil
	}

	clone, err := erc.getInstanceFromGoStruct(registeredPtrTypeRecord.shallowCopyInternalPointer())
	if err != nil {
		return nil, err
	}

	clone.RegisteredPtrTypeRecord().count.value += 1
	clone.RegisteredPtrTypeRecord().deleteScheduled = false
	return clone, nil
}

func (erc *classType) delete(ctx context.Context, handle IEmvalClassBase) error {
	registeredPtrTypeRecord := handle.RegisteredPtrTypeRecord()
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

func (erc *classType) getInstanceFromGoStruct(record *registeredPointerTypeRecord) (IEmvalClassBase, error) {
	if !erc.hasGoStruct {
		return nil, fmt.Errorf("no Go struct registered for class %s", erc.name)
	}

	// @todo: create new instance of Go struct here.
	classHandle := &EmvalClassBase{
		ptr:                     record.ptr,
		ptrType:                 record.ptrType,
		registeredPtrTypeRecord: record,
	}

	return classHandle, nil
}

type EmvalClassBase struct {
	classType               *classType
	ptr                     uint32
	ptrType                 *registeredPointerType
	registeredPtrTypeRecord *registeredPointerTypeRecord
}

func (ecb *EmvalClassBase) ClassType() *classType {
	return ecb.classType
}

func (ecb *EmvalClassBase) Ptr() uint32 {
	return ecb.ptr
}

func (ecb *EmvalClassBase) PtrType() *registeredPointerType {
	return ecb.ptrType
}

func (ecb *EmvalClassBase) RegisteredPtrTypeRecord() *registeredPointerTypeRecord {
	return ecb.registeredPtrTypeRecord
}

type IEmvalClassBase interface {
	ClassType() *classType
	Ptr() uint32
	PtrType() *registeredPointerType
	RegisteredPtrTypeRecord() *registeredPointerTypeRecord
}
