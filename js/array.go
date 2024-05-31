package js

import (
	"fmt"
	"unsafe"
)

type ArrayBuffer struct {
	Buffer []uint8
	Length int
}

type Int8Array struct {
	Buffer []int8
	Length int
}

type Int16Array struct {
	Buffer []int16
	Length int
}

type Uint8Array struct {
	Buffer []uint8
	Length int
}

type Uint16Array struct {
	Buffer []uint16
	Length int
}

type Int32Array struct {
	Buffer []int32
	Length int
}

type Uint32Array struct {
	Buffer []uint32
	Length int
}

type Float32Array struct {
	Buffer []float32
	Length int
}

func (obj *Float32Array) New(argTypes []string, args ...any) (any, error) {
	if len(argTypes) == 3 {
		switch v := args[0].(type) {
		case Uint8Array:
			sourceArray := unsafe.Slice((*float32)(unsafe.Pointer(&v.Buffer[0])), v.Length)
			return &Float32Array{
				Buffer: sourceArray[args[1].(int):(args[2].(int) * 4)],
				Length: args[2].(int) / 4,
			}, nil
		case *Uint8Array:
			sourceArray := unsafe.Slice((*float32)(unsafe.Pointer(&v.Buffer[0])), v.Length)
			return &Float32Array{
				Buffer: sourceArray[args[1].(int):(args[2].(int) * 4)],
				Length: args[2].(int) / 4,
			}, nil
		case ArrayBuffer:
			sourceArray := unsafe.Slice((*float32)(unsafe.Pointer(&v.Buffer[0])), v.Length)
			return &Float32Array{
				Buffer: sourceArray[args[1].(int):(args[2].(int) * 4)],
				Length: args[2].(int) / 4,
			}, nil
		case *ArrayBuffer:
			sourceArray := unsafe.Slice((*float32)(unsafe.Pointer(&v.Buffer[0])), v.Length)
			return &Float32Array{
				Buffer: sourceArray[args[1].(int):(args[2].(int) * 4)],
				Length: args[2].(int) / 4,
			}, nil
		default:
			return nil, fmt.Errorf("unknown source type: %T", v)
		}
	}

	// @todo: implement other constructors.
	return nil, fmt.Errorf("unknown constructor of length %d", len(argTypes))
}

func (obj *Float32Array) Set(input any) error {
	switch v := input.(type) {
	case Float32Array:
		obj.Buffer = v.Buffer
		obj.Length = v.Length
		return nil
	case *Float32Array:
		obj.Buffer = v.Buffer
		obj.Length = v.Length
		return nil
	}

	return fmt.Errorf("set can only receive Float32Array or *Float32Array, got %T", input)
}

type Float64Array struct {
	Buffer []float64
	Length int
}

func (obj *Float64Array) New(argTypes []string, args ...any) (any, error) {
	// @todo: implement me.
	return Float64Array{}, nil
}
