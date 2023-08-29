package embind

import (
	"context"
	"fmt"
	"github.com/tetratelabs/wazero/api"
)

type arrayType struct {
	baseType
	reg            *registeredTuple
	elementsLength int
}

func (at *arrayType) FromWireType(ctx context.Context, mod api.Module, ptr uint64) (any, error) {
	var err error
	rv := make([]any, at.elementsLength)
	for i := 0; i < at.elementsLength; i++ {
		rv[i], err = at.reg.elements[i].read(ctx, mod, api.DecodeI32(ptr))
		if err != nil {
			return nil, err
		}
	}

	_, err = at.reg.rawDestructor.Call(ctx, ptr)
	if err != nil {
		return nil, err
	}

	return rv, nil
}

func (at *arrayType) ToWireType(ctx context.Context, mod api.Module, destructors *[]*destructorFunc, o any) (uint64, error) {
	arr, ok := o.([]any)
	if !ok {
		return 0, fmt.Errorf("incorrect input, not an array, make sure that the input is of type []any")
	}

	if at.elementsLength != len(arr) {
		return 0, fmt.Errorf("incorrect number of tuple elements for %s: expected=%d, actual=%d", at.reg.name, at.elementsLength, len(arr))
	}

	res, err := at.reg.rawConstructor.Call(ctx)
	if err != nil {
		return 0, err
	}

	ptr := res[0]

	for i := 0; i < at.elementsLength; i++ {
		err = at.reg.elements[i].write(ctx, mod, api.DecodeI32(ptr), arr[i])
		if err != nil {
			return 0, err
		}
	}

	if destructors != nil {
		destructorsRef := *destructors
		destructorsRef = append(destructorsRef, &destructorFunc{
			apiFunction: at.reg.rawDestructor,
			args:        []uint64{ptr},
		})
		*destructors = destructorsRef
	}
	return ptr, nil
}

func (at *arrayType) ReadValueFromPointer(ctx context.Context, mod api.Module, pointer uint32) (any, error) {
	ptr, ok := mod.Memory().ReadUint32Le(pointer)
	if !ok {
		return nil, fmt.Errorf("could not read pointer")
	}
	return at.FromWireType(ctx, mod, api.EncodeU32(ptr))
}

func (at *arrayType) HasDestructorFunction() bool {
	return true
}

func (at *arrayType) DestructorFunction(ctx context.Context, mod api.Module, pointer uint32) (*destructorFunc, error) {
	return &destructorFunc{
		apiFunction: at.reg.rawDestructor,
		args:        []uint64{api.EncodeU32(pointer)},
	}, nil
}

func (at *arrayType) GoType() string {
	return "[]any"
}
