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

func (ev *enumValue) Name() string {
	return ev.name
}

func (ev *enumValue) Value() any {
	return ev.cppValue
}

type IEnum interface {
	Name() string
	Type() IType
	Values() []IEnumValue
}

type IEnumValue interface {
	Name() string
	Value() any
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

func (et *enumType) FromWireType(ctx context.Context, mod api.Module, value uint64) (any, error) {
	val, err := et.intHelper.FromWireType(ctx, mod, value)
	if err != nil {
		return nil, err
	}

	return et.mapToGoEnum(val)
}

func (et *enumType) ToWireType(ctx context.Context, mod api.Module, destructors *[]*destructorFunc, o any) (uint64, error) {
	val, ok := et.valuesByGoValue[o]
	if !ok {
		val, ok = et.valuesByCppValue[o]
		if !ok {
			return 0, fmt.Errorf("could not map enum value %v, enum value not registered as Go or C++ enum", o)
		}
	}

	return et.intHelper.ToWireType(ctx, mod, destructors, val.cppValue)
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
		return value, nil
	}

	val, ok := et.valuesByCppValue[value]
	if !ok {
		return nil, fmt.Errorf("could not map enum value %v, enum value not registered as C++ enum", value)
	}

	if !val.hasGoValue {
		return value, nil
	}

	return val.goValue, nil
}

func (et *enumType) GoType() string {
	// @todo: use Go name when registered?
	return et.name
}

func (et *enumType) Type() IType {
	return &exposedType{registeredType: &et.intHelper}
}

func (et *enumType) Name() string {
	return et.name
}

func (et *enumType) Values() []IEnumValue {
	values := make([]IEnumValue, 0)
	for i := range et.valuesByName {
		values = append(values, et.valuesByName[i])
	}
	return values
}

func (e *engine) GetEnums() []IEnum {
	enums := make([]IEnum, 0)
	for i := range e.registeredEnums {
		enums = append(enums, e.registeredEnums[i])
	}
	return enums
}
