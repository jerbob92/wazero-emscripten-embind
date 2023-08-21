package embind

import (
	"context"
	"fmt"

	"github.com/tetratelabs/wazero/api"
)

type intType struct {
	baseType
	size   int32
	signed bool
}

// @todo: implement min/max checks?

func (it *intType) FromWireType(ctx context.Context, mod api.Module, value uint64) (any, error) {
	if it.size == 1 {
		if !it.signed {
			return uint8(api.DecodeI32(value)), nil
		}

		return int8(api.DecodeI32(value)), nil
	} else if it.size == 2 {
		if !it.signed {
			return uint16(api.DecodeI32(value)), nil
		}

		return int16(api.DecodeI32(value)), nil
	} else if it.size == 4 {
		if !it.signed {
			return api.DecodeU32(value), nil
		}

		return api.DecodeI32(value), nil
	}

	return nil, fmt.Errorf("unknown integer size")
}

func (it *intType) ToWireType(ctx context.Context, mod api.Module, destructors *[]*destructorFunc, o any) (uint64, error) {
	if it.size == 1 {
		if !it.signed {
			uint8Val, ok := o.(uint8)
			if ok {
				return uint64(uint8Val), nil
			}

			return 0, fmt.Errorf("value must be of type uint8, is %T", o)
		}

		int8Val, ok := o.(int8)
		if ok {
			return uint64(int8Val), nil
		}

		return 0, fmt.Errorf("value must be of type int8, is %T", o)
	} else if it.size == 2 {
		if !it.signed {
			uint16Val, ok := o.(uint16)
			if ok {
				return uint64(uint16Val), nil
			}

			return 0, fmt.Errorf("value must be of type uint16, is %T", o)
		}

		int16Val, ok := o.(int16)
		if ok {
			return uint64(int16Val), nil
		}

		return 0, fmt.Errorf("value must be of type int16, is %T", o)
	} else if it.size == 4 {
		if !it.signed {
			uint32Val, ok := o.(uint32)
			if ok {
				return api.EncodeU32(uint32Val), nil
			}

			return 0, fmt.Errorf("value must be of type uint32, is %T", o)
		}

		int32Val, ok := o.(int32)
		if ok {
			return api.EncodeI32(int32Val), nil
		}

		return 0, fmt.Errorf("value must be of type int32, is %T", o)
	}

	return 0, fmt.Errorf("unknown integer size for %T", o)
}

func (it *intType) ReadValueFromPointer(ctx context.Context, mod api.Module, pointer uint32) (any, error) {
	if it.size == 1 {
		val, _ := mod.Memory().ReadByte(pointer)
		if !it.signed {
			return uint8(val), nil
		}
		return int8(val), nil
	} else if it.size == 2 {
		val, _ := mod.Memory().ReadUint16Le(pointer)
		if !it.signed {
			return uint16(val), nil
		}
		return int16(val), nil
	} else if it.size == 4 {
		val, _ := mod.Memory().ReadUint32Le(pointer)
		if !it.signed {
			return uint32(val), nil
		}
		return int32(val), nil
	}

	return nil, fmt.Errorf("unknown integer type: %s", it.name)
}

func (it *intType) GoType() string {
	if it.size == 1 {
		if !it.signed {
			return "uint8"
		}
		return "int8"
	} else if it.size == 2 {
		if !it.signed {
			return "uint16"
		}
		return "int16"
	}

	if !it.signed {
		return "uint32"
	}
	return "int32"
}
