package embind

import (
	"context"
	"fmt"
	"github.com/jerbob92/wazero-emscripten-embind/types"

	"github.com/tetratelabs/wazero/api"
)

type boolType struct {
	baseType
	size     int32
	trueVal  int32
	falseVal int32
}

func (bt *boolType) FromWireType(ctx context.Context, mod api.Module, value uint64) (any, error) {
	// ambiguous emscripten ABI: sometimes return values are
	// true or false, and sometimes integers (0 or 1)
	return value > 0, nil
}

func (bt *boolType) ToWireType(ctx context.Context, mod api.Module, destructors *[]*destructorFunc, o any) (uint64, error) {
	if o == nil || o == types.Undefined {
		return api.EncodeI32(bt.falseVal), nil
	}

	val, ok := o.(bool)
	if ok {
		if val {
			return api.EncodeI32(bt.trueVal), nil
		}
		return api.EncodeI32(bt.falseVal), nil
	}

	stringVal, ok := o.(string)
	if ok {
		if stringVal != "" {
			return api.EncodeI32(bt.trueVal), nil
		}
		return api.EncodeI32(bt.falseVal), nil
	}

	// Float64 is big enough for any number.
	numberVal := float64(0)
	hasNumberVal := false
	switch v := o.(type) {
	case int:
		numberVal = float64(v)
		hasNumberVal = true
	case uint:
		numberVal = float64(v)
		hasNumberVal = true
	case int8:
		numberVal = float64(v)
		hasNumberVal = true
	case uint8:
		numberVal = float64(v)
		hasNumberVal = true
	case int16:
		numberVal = float64(v)
		hasNumberVal = true
	case uint16:
		numberVal = float64(v)
		hasNumberVal = true
	case int32:
		numberVal = float64(v)
		hasNumberVal = true
	case uint32:
		numberVal = float64(v)
		hasNumberVal = true
	case int64:
		numberVal = float64(v)
		hasNumberVal = true
	case uint64:
		numberVal = float64(v)
		hasNumberVal = true
	case float32:
		numberVal = float64(v)
		hasNumberVal = true
	case float64:
		numberVal = v
		hasNumberVal = true
	}

	if hasNumberVal {
		if numberVal > 0 {
			return api.EncodeI32(bt.trueVal), nil
		}
		return api.EncodeI32(bt.falseVal), nil
	}

	// @todo: implement nil pointer check

	// Any other type could be considered true?
	return api.EncodeI32(bt.trueVal), nil
}

func (bt *boolType) ReadValueFromPointer(ctx context.Context, mod api.Module, pointer uint32) (any, error) {
	if bt.size == 1 {
		val, _ := mod.Memory().ReadByte(pointer)
		return bt.FromWireType(ctx, mod, uint64(val))
	} else if bt.size == 2 {
		val, _ := mod.Memory().ReadUint16Le(pointer)
		return bt.FromWireType(ctx, mod, uint64(val))
	} else if bt.size == 4 {
		val, _ := mod.Memory().ReadUint32Le(pointer)
		return bt.FromWireType(ctx, mod, uint64(val))
	} else {
		return nil, fmt.Errorf("unknown boolean type size %d: %s", bt.size, bt.name)
	}
}

func (bt *boolType) GoType() string {
	return "bool"
}
