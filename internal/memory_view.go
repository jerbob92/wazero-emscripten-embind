package embind

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/tetratelabs/wazero/api"
)

type memoryViewType struct {
	baseType
	dataTypeIndex int32
	nativeType    any
	nativeSize    uint32
}

func (mvt *memoryViewType) FromWireType(ctx context.Context, mod api.Module, value uint64) (any, error) {
	memoryViewPtr := api.DecodeU32(value)

	var ok bool

	size, ok := mod.Memory().ReadUint32Le(memoryViewPtr)
	if !ok {
		return nil, fmt.Errorf("could not read size of memory view")
	}

	pointer, ok := mod.Memory().ReadUint32Le(memoryViewPtr + 4)
	if !ok {
		return nil, fmt.Errorf("could not read pointer of memory view")
	}

	var typedMemoryView any

	if mvt.dataTypeIndex == 0 {
		typedMemoryView, ok = memoryAs[int8](mod.Memory(), pointer, size/mvt.nativeSize, size)
	} else if mvt.dataTypeIndex == 1 {
		typedMemoryView, ok = memoryAs[uint8](mod.Memory(), pointer, size/mvt.nativeSize, size)
	} else if mvt.dataTypeIndex == 2 {
		typedMemoryView, ok = memoryAs[int16](mod.Memory(), pointer, size/mvt.nativeSize, size)
	} else if mvt.dataTypeIndex == 3 {
		typedMemoryView, ok = memoryAs[uint16](mod.Memory(), pointer, size/mvt.nativeSize, size)
	} else if mvt.dataTypeIndex == 4 {
		typedMemoryView, ok = memoryAs[int32](mod.Memory(), pointer, size/mvt.nativeSize, size)
	} else if mvt.dataTypeIndex == 5 {
		typedMemoryView, ok = memoryAs[uint32](mod.Memory(), pointer, size/mvt.nativeSize, size)
	} else if mvt.dataTypeIndex == 6 {
		typedMemoryView, ok = memoryAs[float32](mod.Memory(), pointer, size/mvt.nativeSize, size)
	} else if mvt.dataTypeIndex == 7 {
		typedMemoryView, ok = memoryAs[float64](mod.Memory(), pointer, size/mvt.nativeSize, size)
	} else if mvt.dataTypeIndex == 8 {
		typedMemoryView, ok = memoryAs[int64](mod.Memory(), pointer, size/mvt.nativeSize, size)
	} else if mvt.dataTypeIndex == 9 {
		typedMemoryView, ok = memoryAs[uint64](mod.Memory(), pointer, size/mvt.nativeSize, size)
	} else {
		return nil, fmt.Errorf("unknown memory view type %s", mvt.name)
	}

	if !ok {
		return nil, fmt.Errorf("could not create memory view")
	}

	return typedMemoryView, nil
}

func memoryAs[T any](memory api.Memory, offset uint32, size uint32, length uint32) ([]T, bool) {
	memoryView, ok := memory.Read(offset, length)
	if !ok {
		return nil, ok
	}

	return unsafe.Slice((*T)(unsafe.Pointer(&memoryView[0])), size), true
}

func (mvt *memoryViewType) ToWireType(ctx context.Context, mod api.Module, destructors *[]*destructorFunc, o any) (uint64, error) {
	return 0, fmt.Errorf("ToWireType is not supported for memory views")
}

func (mvt *memoryViewType) ReadValueFromPointer(ctx context.Context, mod api.Module, pointer uint32) (any, error) {
	return mvt.FromWireType(ctx, mod, api.EncodeU32(pointer))
}

func (mvt *memoryViewType) GoType() string {
	if mvt.dataTypeIndex == 0 {
		return "[]int8"
	} else if mvt.dataTypeIndex == 1 {
		return "[]uint8"
	} else if mvt.dataTypeIndex == 2 {
		return "[]int16"
	} else if mvt.dataTypeIndex == 3 {
		return "[]uint16"
	} else if mvt.dataTypeIndex == 4 {
		return "[]int32"
	} else if mvt.dataTypeIndex == 5 {
		return "[]uint32"
	} else if mvt.dataTypeIndex == 6 {
		return "[]float32"
	} else if mvt.dataTypeIndex == 7 {
		return "[]float64"
	} else if mvt.dataTypeIndex == 8 {
		return "[]int64"
	} else if mvt.dataTypeIndex == 9 {
		return "[]uint64"
	}

	return "[]uint64"
}
