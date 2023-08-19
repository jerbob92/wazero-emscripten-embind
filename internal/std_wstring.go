package embind

import (
	"context"
	"fmt"
	"github.com/tetratelabs/wazero/api"
)

type stdWStringType struct {
	baseType
	charSize int32
}

// @todo: decide whether we should return string, []byte or something else for UTF16/UTF32.

func (swst *stdWStringType) FromWireType(ctx context.Context, mod api.Module, value uint64) (any, error) {
	strPointer := api.DecodeI32(value)

	length, ok := mod.Memory().ReadUint32Le(uint32(strPointer))
	if !ok {
		return nil, fmt.Errorf("could not read length of string")
	}

	payload := strPointer + 4

	data, ok := mod.Memory().Read(uint32(payload), length*uint32(swst.charSize))
	if !ok {
		return nil, fmt.Errorf("could not read data of string")
	}

	str := string(data)

	_, err := mod.ExportedFunction("free").Call(ctx, value)
	if err != nil {
		return nil, err
	}

	return str, nil
}

func (swst *stdWStringType) ToWireType(ctx context.Context, mod api.Module, destructors *[]*destructorFunc, o any) (uint64, error) {
	// @todo: implement me.
	return 0, nil
}

func (swst *stdWStringType) ReadValueFromPointer(ctx context.Context, mod api.Module, pointer uint32) (any, error) {
	ptr, ok := mod.Memory().ReadUint32Le(pointer)
	if !ok {
		return nil, fmt.Errorf("could not read pointer")
	}
	return swst.FromWireType(ctx, mod, api.EncodeU32(ptr))
}

func (swst *stdWStringType) HasDestructorFunction() bool {
	return true
}

func (swst *stdWStringType) DestructorFunction(ctx context.Context, mod api.Module, pointer uint32) (*destructorFunc, error) {
	return &destructorFunc{
		apiFunction: mod.ExportedFunction("free"),
		args:        []uint64{api.EncodeU32(pointer)},
	}, nil
}

func (swst *stdWStringType) GoType() string {
	return "string"
}
