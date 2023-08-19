package embind

import (
	"context"
	"fmt"
	"github.com/tetratelabs/wazero/api"
)

type bigintType struct {
	baseType
	size   int32
	signed bool
}

// @todo: implement min/max checks?

func (bt *bigintType) FromWireType(ctx context.Context, mod api.Module, value uint64) (any, error) {
	if bt.size == 8 {
		if !bt.signed {
			return uint64(value), nil
		}

		return int64(value), nil
	}

	return nil, fmt.Errorf("unknown bigint size")
}

func (bt *bigintType) ToWireType(ctx context.Context, mod api.Module, destructors *[]*destructorFunc, o any) (uint64, error) {
	if bt.size == 8 {
		if !bt.signed {
			uint64Val, ok := o.(uint64)
			if ok {
				return uint64(uint64Val), nil
			}

			return 0, fmt.Errorf("value must be of type uint64")
		}

		int64Val, ok := o.(int64)
		if ok {
			return uint64(int64Val), nil
		}

		return 0, fmt.Errorf("value must be of type int64")
	}

	return 0, fmt.Errorf("unknown bigint size")
}

func (bt *bigintType) ReadValueFromPointer(ctx context.Context, mod api.Module, pointer uint32) (any, error) {
	if bt.size == 8 {
		val, _ := mod.Memory().ReadUint64Le(pointer)
		if !bt.signed {
			return uint64(val), nil
		}
		return int64(val), nil
	}

	return nil, fmt.Errorf("unknown bigint type: %s", bt.name)
}

func (bt *bigintType) NativeType() api.ValueType {
	return api.ValueTypeI64
}
