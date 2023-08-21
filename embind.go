package embind

import (
	internal "github.com/jerbob92/wazero-emscripten-embind/internal"
	"github.com/tetratelabs/wazero"
)

type Engine interface {
	internal.IEngine
	NewFunctionExporterForModule(guest wazero.CompiledModule) FunctionExporter
}

type Enum interface {
	internal.Enum
}

type EngineKey = internal.EngineKey

func CreateEngine(config internal.IEngineConfig) Engine {
	return &wazeroEngine{
		config:  config,
		IEngine: internal.CreateEngine(config),
	}
}

func NewConfig() internal.IEngineConfig {
	return &internal.EngineConfig{}
}

type EmvalConstructor interface {
	internal.IEmvalConstructor
}

type EmvalFunctionMapper interface {
	internal.IEmvalFunctionMapper
}

type EmvalClassBase interface {
	internal.IEmvalClassBase
}
