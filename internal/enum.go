package embind

import (
	"context"
	"fmt"
	"github.com/tetratelabs/wazero/api"
)

type Enum interface {
	Type() any
	Values() map[string]any
}

type enumValue struct {
	name        string
	hasCppValue bool
	cppValue    any
	rawCppValue uint64
	hasGoValue  bool
	goValue     any
}

func (ev *enumValue) validate() error {
	if ev.hasGoValue && ev.hasCppValue {
		if ev.goValue != ev.cppValue {
			return fmt.Errorf("enum value %s has a different value in Go than in C++ (go: %v (%T), cpp: %v (%T))", ev.name, ev.goValue, ev.goValue, ev.cppValue, ev.cppValue)
		}
	}

	return nil
}

type enumType struct {
	baseType
	intHelper        intType // Enums are basically ints, we use the intType underwater to make things easier.
	valuesByName     map[string]*enumValue
	valuesByCppValue map[any]*enumValue
	valuesByGoValue  map[any]*enumValue
	registeredInGo   bool
	goValue          any
}

func (et *enumType) validate() error {
	for i := range et.valuesByName {
		if err := et.valuesByName[i].validate(); err != nil {
			return fmt.Errorf("error while validating enum %s: %w", et.name, err)
		}
	}

	return nil
}

func (et *enumType) FromWireType(ctx context.Context, mod api.Module, value uint64) (any, error) {
	val, err := et.intHelper.FromWireType(ctx, mod, value)
	if err != nil {
		return nil, err
	}

	return et.mapToGoEnum(val)
}

func (et *enumType) ToWireType(ctx context.Context, mod api.Module, destructors *[]*destructorFunc, o any) (uint64, error) {
	if !et.registeredInGo {
		return 0, fmt.Errorf("could not map enum value %v, enum not registered as Go enum", o)
	}

	val, ok := et.valuesByGoValue[o]
	if !ok {
		return 0, fmt.Errorf("could not map enum value %v, enum value not registered as Go enum", o)
	}

	return et.intHelper.ToWireType(ctx, mod, destructors, val.goValue)
}

func (et *enumType) ReadValueFromPointer(ctx context.Context, mod api.Module, pointer uint32) (any, error) {
	val, err := et.intHelper.ReadValueFromPointer(ctx, mod, pointer)
	if err != nil {
		return nil, err
	}

	return et.mapToGoEnum(val)
}

func (et *enumType) mapToGoEnum(value any) (any, error) {
	if !et.registeredInGo {
		return nil, fmt.Errorf("could not map enum value %v, enum not registered as Go enum", value)
	}

	val, ok := et.valuesByCppValue[value]
	if !ok {
		return nil, fmt.Errorf("could not map enum value %v, enum value not registered as C++ enum", value)
	}

	if !val.hasGoValue {
		return nil, fmt.Errorf("could not map enum value %v, enum value has no registered Go value", value)
	}

	return val.goValue, nil
}

func (et *enumType) GoType() string {
	// @todo: use Go name when registered?
	return et.name
}
