package embind

import (
	"context"
	"fmt"
	"github.com/tetratelabs/wazero/api"
)

type objectType struct {
	baseType
	reg            *registeredObject
	elementsLength int
}

func (ot *objectType) FromWireType(ctx context.Context, mod api.Module, ptr uint64) (any, error) {
	var err error
	rv := map[string]any{}

	for i := range ot.reg.fields {
		rv[ot.reg.fields[i].fieldName], err = ot.reg.fields[i].read(ctx, mod, api.DecodeI32(ptr))
	}

	_, err = ot.reg.rawDestructor.Call(ctx, ptr)
	if err != nil {
		return nil, err
	}

	return rv, nil
}

func (ot *objectType) ToWireType(ctx context.Context, mod api.Module, destructors *[]*destructorFunc, o any) (uint64, error) {
	obj, ok := o.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("incorrect input, not an string map")
	}

	for i := range ot.reg.fields {
		if _, ok = obj[ot.reg.fields[i].fieldName]; !ok {
			return 0, fmt.Errorf("missing field: %s", ot.reg.fields[i].fieldName)
		}
	}

	res, err := ot.reg.rawConstructor.Call(ctx)
	if err != nil {
		return 0, err
	}

	ptr := res[0]

	for i := range ot.reg.fields {
		err = ot.reg.fields[i].write(ctx, mod, api.DecodeI32(ptr), obj[ot.reg.fields[i].fieldName])
		if err != nil {
			return 0, err
		}
	}

	if destructors != nil {
		destructorsRef := *destructors
		destructorsRef = append(destructorsRef, &destructorFunc{
			apiFunction: ot.reg.rawDestructor,
			args:        []uint64{ptr},
		})
		*destructors = destructorsRef
	}
	return ptr, nil
}

func (ot *objectType) ReadValueFromPointer(ctx context.Context, mod api.Module, pointer uint32) (any, error) {
	ptr, ok := mod.Memory().ReadUint32Le(pointer)
	if !ok {
		return nil, fmt.Errorf("could not read pointer")
	}
	return ot.FromWireType(ctx, mod, api.EncodeU32(ptr))
}

func (ot *objectType) HasDestructorFunction() bool {
	return true
}

func (ot *objectType) DestructorFunction(ctx context.Context, mod api.Module, pointer uint32) (*destructorFunc, error) {
	return &destructorFunc{
		apiFunction: ot.reg.rawDestructor,
		args:        []uint64{api.EncodeU32(pointer)},
	}, nil
}

func (ot *objectType) GoType() string {
	return "map[string]any"
}
