package embind

import (
	"context"
	"fmt"
	"github.com/tetratelabs/wazero/api"
)

type floatType struct {
	baseType
	size int32
}

func (ft *floatType) FromWireType(ctx context.Context, mod api.Module, value uint64) (any, error) {
	if ft.size == 4 {
		return api.DecodeF32(value), nil
	}
	if ft.size == 8 {
		return api.DecodeF64(value), nil
	}
	return nil, fmt.Errorf("unknown float size")
}

func (ft *floatType) ToWireType(ctx context.Context, mod api.Module, destructors *[]*destructorFunc, o any) (uint64, error) {
	if ft.size == 4 {
		f32Val, ok := o.(float32)
		if ok {
			return api.EncodeF32(f32Val), nil
		}

		return 0, fmt.Errorf("value must be of type float32")
	}

	if ft.size == 8 {
		f64Val, ok := o.(float64)
		if ok {
			return api.EncodeF64(f64Val), nil
		}

		return 0, fmt.Errorf("value must be of type float64")
	}

	return 0, fmt.Errorf("unknown float size")
}

func (ft *floatType) ReadValueFromPointer(ctx context.Context, mod api.Module, pointer uint32) (any, error) {
	if ft.size == 4 {
		val, _ := mod.Memory().ReadFloat32Le(pointer)
		return val, nil
	} else if ft.size == 8 {
		val, _ := mod.Memory().ReadUint64Le(pointer)
		return val, nil
	}

	return nil, fmt.Errorf("unknown float type: %s", ft.name)
}

func (ft *floatType) NativeType() api.ValueType {
	if ft.size == 4 {
		return api.ValueTypeF32
	}

	return api.ValueTypeF64
}

func (ft *floatType) GoType() string {
	if ft.size == 4 {
		return "float32"
	}

	return "float64"
}
