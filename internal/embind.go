package embind

import (
	"context"
	"fmt"
	"github.com/tetratelabs/wazero/api"
)

type IEngine interface {
	Attach(ctx context.Context) context.Context
	CallPublicSymbol(ctx context.Context, name string, arguments ...any) (any, error)
	RegisterConstant(name string, val any) error
	RegisterEnum(name string, enum Enum) error
	RegisterClass(name string, class any) error
	RegisterEmvalSymbol(name string, symbol any) error
	EmvalToHandle(value any) int32
	EmvalToValue(handle int32) (any, error)
}

type IEngineConfig interface {
}

type EngineConfig struct {
}

func GetEngineFromContext(ctx context.Context) (IEngine, error) {
	raw := ctx.Value(EngineKey{})
	if raw == nil {
		return nil, fmt.Errorf("embind engine not found in context")
	}

	value, ok := raw.(IEngine)
	if !ok {
		return nil, fmt.Errorf("context value %v not of type %T", value, new(IEngine))
	}

	return value, nil
}

func MustGetEngineFromContext(ctx context.Context, mod api.Module) IEngine {
	e, err := GetEngineFromContext(ctx)
	if err != nil {
		panic(fmt.Errorf("could not get embind engine from context: %w, make sure to create an engine with embind.CreateEngine() and to attach it to the context with \"ctx = context.WithValue(ctx, embind.EngineKey{}, engine)\"", err))
	}

	if e.(*engine).mod != nil {
		if mod != nil && e.(*engine).mod != mod {
			panic(fmt.Errorf("could not get embind engine from context, this engine was created for another Wazero api.Module"))
		}
	}

	if mod != nil {
		// Make sure we have the api module set.
		e.(*engine).mod = mod
	}

	return e
}

// EngineKey Use this key to add the engine to your context:
// ctx = context.WithValue(ctx, embind.EngineKey{}, engine)
type EngineKey struct{}

// CreateEngine returns a new embind engine to attach to your context.
// Be sure to attach it before you run InstantiateModule on the runtime, unless
// you run the _start/_initialize function manually.
func CreateEngine(config IEngineConfig) IEngine {
	return &engine{
		config:               config,
		publicSymbols:        map[string]*publicSymbol{},
		registeredTypes:      map[int32]registeredType{},
		typeDependencies:     map[int32][]int32{},
		awaitingDependencies: map[int32][]*awaitingDependency{},
		registeredConstants:  map[string]*registeredConstant{},
		registeredEnums:      map[string]*enumType{},
		registeredClasses:    map[string]*classType{},
		registeredPointers:   map[int32]*registeredPointer{},
		registeredTuples:     map[int32]*registeredTuple{},
		registeredObjects:    map[int32]*registeredObject{},
		emvalEngine:          createEmvalEngine(),
	}
}
