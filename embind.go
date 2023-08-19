package embind

import (
	internal "github.com/jerbob92/wazero-emscripten-embind/internal"
)

type Engine interface {
	internal.Engine
}

type Enum interface {
	internal.Enum
}

type EngineKey = internal.EngineKey

func CreateEngine() Engine {
	return internal.CreateEngine()
}

type EmvalConstructor interface {
	internal.EmvalConstructor
}

type EmvalFunctionMapper interface {
	internal.EmvalFunctionMapper
}

type EmvalClassBase interface {
	internal.IEmvalClassBase
}
