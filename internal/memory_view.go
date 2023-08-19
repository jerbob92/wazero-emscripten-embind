package embind

import (
	"context"
	"fmt"
	"github.com/tetratelabs/wazero/api"
	"unsafe"
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
		typedMemoryView, ok = memoryAs[int8](mod.Memory(), pointer, size, size*mvt.nativeSize)
	} else if mvt.dataTypeIndex == 1 {
		typedMemoryView, ok = memoryAs[uint8](mod.Memory(), pointer, size, size*mvt.nativeSize)
	} else if mvt.dataTypeIndex == 2 {
		typedMemoryView, ok = memoryAs[int16](mod.Memory(), pointer, size, size*mvt.nativeSize)
	} else if mvt.dataTypeIndex == 3 {
		typedMemoryView, ok = memoryAs[uint16](mod.Memory(), pointer, size, size*mvt.nativeSize)
	} else if mvt.dataTypeIndex == 4 {
		typedMemoryView, ok = memoryAs[int32](mod.Memory(), pointer, size, size*mvt.nativeSize)
	} else if mvt.dataTypeIndex == 5 {
		typedMemoryView, ok = memoryAs[uint8](mod.Memory(), pointer, size, size*mvt.nativeSize)
	} else if mvt.dataTypeIndex == 6 {
		typedMemoryView, ok = memoryAs[float32](mod.Memory(), pointer, size, size*mvt.nativeSize)
	} else if mvt.dataTypeIndex == 7 {
		typedMemoryView, ok = memoryAs[float64](mod.Memory(), pointer, size, size*mvt.nativeSize)
	} else if mvt.dataTypeIndex == 8 {
		typedMemoryView, ok = memoryAs[int64](mod.Memory(), pointer, size, size*mvt.nativeSize)
	} else if mvt.dataTypeIndex == 9 {
		typedMemoryView, ok = memoryAs[uint64](mod.Memory(), pointer, size, size*mvt.nativeSize)
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
