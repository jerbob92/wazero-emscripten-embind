package embind

import (
	"context"

	"github.com/jerbob92/wazero-emscripten-embind/types"

	"github.com/tetratelabs/wazero/api"
)

type voidType struct {
	baseType
}

func (vt *voidType) FromWireType(ctx context.Context, mod api.Module, value uint64) (any, error) {
	return types.Undefined, nil
}

func (vt *voidType) ToWireType(ctx context.Context, mod api.Module, destructors *[]*destructorFunc, o any) (uint64, error) {
	// TODO: assert if anything else is given? (comment from Emscripten)
	return 0, nil
}

func (vt *voidType) NativeType() api.ValueType {
	return 0
}

func (vt *voidType) GoType() string {
	return ""
}
