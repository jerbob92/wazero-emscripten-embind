// Code generated by wazero-emscripten-embind, DO NOT EDIT.
package generated

import (
	"context"

	"github.com/jerbob92/wazero-emscripten-embind"
)

type ClassBase struct {
	embind.ClassBase
}

func (class *ClassBase) Invoke(ctx context.Context, arg0 string) error {
	_, err := class.CallMethod(ctx, class, "invoke", arg0)
	return err
}

func ClassBaseStaticExtend(e embind.Engine, ctx context.Context, arg0 string, arg1 any) (any, error) {
	res, err := e.CallStaticClassMethod(ctx, "Base", "extend", arg0, arg1)
	if err != nil {
		return nil, err
	}

	return res.(any), nil
}

func ClassBaseStaticImplement(e embind.Engine, ctx context.Context, arg0 any) (*ClassBaseWrapper, error) {
	res, err := e.CallStaticClassMethod(ctx, "Base", "implement", arg0)
	if err != nil {
		return nil, err
	}

	return res.(*ClassBaseWrapper), nil
}

type ClassBaseWrapper struct {
	embind.ClassBase
}

func (class *ClassBaseWrapper) NotifyOnDestruction(ctx context.Context) error {
	_, err := class.CallMethod(ctx, class, "notifyOnDestruction")
	return err
}

func ClassBaseWrapperStaticExtend(e embind.Engine, ctx context.Context, arg0 string, arg1 any) (any, error) {
	res, err := e.CallStaticClassMethod(ctx, "BaseWrapper", "extend", arg0, arg1)
	if err != nil {
		return nil, err
	}

	return res.(any), nil
}

func ClassBaseWrapperStaticImplement(e embind.Engine, ctx context.Context, arg0 any) (*ClassBaseWrapper, error) {
	res, err := e.CallStaticClassMethod(ctx, "BaseWrapper", "implement", arg0)
	if err != nil {
		return nil, err
	}

	return res.(*ClassBaseWrapper), nil
}

type ClassC struct {
	embind.ClassBase
}

func NewClassC(e embind.Engine, ctx context.Context) (*ClassC, error) {
	res, err := e.CallPublicSymbol(ctx, "C")
	if err != nil {
		return nil, err
	}

	return res.(*ClassC), nil
}

type ClassDerived struct {
	embind.ClassBase
}

func ClassDerivedStaticExtend(e embind.Engine, ctx context.Context, arg0 string, arg1 any) (any, error) {
	res, err := e.CallStaticClassMethod(ctx, "Derived", "extend", arg0, arg1)
	if err != nil {
		return nil, err
	}

	return res.(any), nil
}

func ClassDerivedStaticImplement(e embind.Engine, ctx context.Context, arg0 any) (*ClassBaseWrapper, error) {
	res, err := e.CallStaticClassMethod(ctx, "Derived", "implement", arg0)
	if err != nil {
		return nil, err
	}

	return res.(*ClassBaseWrapper), nil
}

type ClassInterface struct {
	embind.ClassBase
}

func (class *ClassInterface) Invoke(ctx context.Context, arg0 string) error {
	_, err := class.CallMethod(ctx, class, "invoke", arg0)
	return err
}

func ClassInterfaceStaticExtend(e embind.Engine, ctx context.Context, arg0 string, arg1 any) (any, error) {
	res, err := e.CallStaticClassMethod(ctx, "Interface", "extend", arg0, arg1)
	if err != nil {
		return nil, err
	}

	return res.(any), nil
}

func ClassInterfaceStaticImplement(e embind.Engine, ctx context.Context, arg0 any) (*ClassInterfaceWrapper, error) {
	res, err := e.CallStaticClassMethod(ctx, "Interface", "implement", arg0)
	if err != nil {
		return nil, err
	}

	return res.(*ClassInterfaceWrapper), nil
}

type ClassInterfaceWrapper struct {
	embind.ClassBase
}

func (class *ClassInterfaceWrapper) NotifyOnDestruction(ctx context.Context) error {
	_, err := class.CallMethod(ctx, class, "notifyOnDestruction")
	return err
}

func ClassInterfaceWrapperStaticExtend(e embind.Engine, ctx context.Context, arg0 string, arg1 any) (any, error) {
	res, err := e.CallStaticClassMethod(ctx, "InterfaceWrapper", "extend", arg0, arg1)
	if err != nil {
		return nil, err
	}

	return res.(any), nil
}

func ClassInterfaceWrapperStaticImplement(e embind.Engine, ctx context.Context, arg0 any) (*ClassInterfaceWrapper, error) {
	res, err := e.CallStaticClassMethod(ctx, "InterfaceWrapper", "implement", arg0)
	if err != nil {
		return nil, err
	}

	return res.(*ClassInterfaceWrapper), nil
}

type ClassMap_int__string_ struct {
	embind.ClassBase
}

func (class *ClassMap_int__string_) Get(ctx context.Context, arg0 int32) (any, error) {
	res, err := class.CallMethod(ctx, class, "get", arg0)
	if err != nil {
		return nil, err
	}

	return res.(any), nil
}

func (class *ClassMap_int__string_) Keys(ctx context.Context) (*ClassVector_int_, error) {
	res, err := class.CallMethod(ctx, class, "keys")
	if err != nil {
		return nil, err
	}

	return res.(*ClassVector_int_), nil
}

func (class *ClassMap_int__string_) Set(ctx context.Context, arg0 int32, arg1 string) error {
	_, err := class.CallMethod(ctx, class, "set", arg0, arg1)
	return err
}

func (class *ClassMap_int__string_) Size(ctx context.Context) (uint32, error) {
	res, err := class.CallMethod(ctx, class, "size")
	if err != nil {
		return uint32(0), err
	}

	return res.(uint32), nil
}

func NewClassMap_int__string_(e embind.Engine, ctx context.Context) (*ClassMap_int__string_, error) {
	res, err := e.CallPublicSymbol(ctx, "map_int__string_")
	if err != nil {
		return nil, err
	}

	return res.(*ClassMap_int__string_), nil
}

type ClassMyClass struct {
	embind.ClassBase
}

func (class *ClassMyClass) GetX(ctx context.Context) (int32, error) {
	res, err := class.GetProperty(ctx, class, "x")
	if err != nil {
		return int32(0), err
	}

	return res.(int32), nil
}
func (class *ClassMyClass) SetX(ctx context.Context, val int32) error {
	return class.SetProperty(ctx, class, "x", val)
}

func (class *ClassMyClass) GetY(ctx context.Context) (string, error) {
	res, err := class.GetProperty(ctx, class, "y")
	if err != nil {
		return "", err
	}

	return res.(string), nil
}

func (class *ClassMyClass) CombineY(ctx context.Context, arg0 string) (string, error) {
	res, err := class.CallMethod(ctx, class, "combineY", arg0)
	if err != nil {
		return "", err
	}

	return res.(string), nil
}

func (class *ClassMyClass) IncrementX0(ctx context.Context) error {
	_, err := class.CallMethod(ctx, class, "incrementX")
	return err
}

func (class *ClassMyClass) IncrementX1(ctx context.Context, arg0 int32) error {
	_, err := class.CallMethod(ctx, class, "incrementX", arg0)
	return err
}

func ClassMyClassStaticGetStringFromInstance(e embind.Engine, ctx context.Context, arg0 *ClassMyClass) (string, error) {
	res, err := e.CallStaticClassMethod(ctx, "MyClass", "getStringFromInstance", arg0)
	if err != nil {
		return "", err
	}

	return res.(string), nil
}

func NewClassMyClass1(e embind.Engine, ctx context.Context, arg0 int32) (*ClassMyClass, error) {
	res, err := e.CallPublicSymbol(ctx, "MyClass", arg0)
	if err != nil {
		return nil, err
	}

	return res.(*ClassMyClass), nil
}

func NewClassMyClass2(e embind.Engine, ctx context.Context, arg0 int32, arg1 string) (*ClassMyClass, error) {
	res, err := e.CallPublicSymbol(ctx, "MyClass", arg0, arg1)
	if err != nil {
		return nil, err
	}

	return res.(*ClassMyClass), nil
}

type ClassVector_int_ struct {
	embind.ClassBase
}

func (class *ClassVector_int_) Get(ctx context.Context, arg0 uint32) (any, error) {
	res, err := class.CallMethod(ctx, class, "get", arg0)
	if err != nil {
		return nil, err
	}

	return res.(any), nil
}

func (class *ClassVector_int_) Push_back(ctx context.Context, arg0 int32) error {
	_, err := class.CallMethod(ctx, class, "push_back", arg0)
	return err
}

func (class *ClassVector_int_) Resize(ctx context.Context, arg0 uint32, arg1 int32) error {
	_, err := class.CallMethod(ctx, class, "resize", arg0, arg1)
	return err
}

func (class *ClassVector_int_) Set(ctx context.Context, arg0 uint32, arg1 int32) (bool, error) {
	res, err := class.CallMethod(ctx, class, "set", arg0, arg1)
	if err != nil {
		return bool(false), err
	}

	return res.(bool), nil
}

func (class *ClassVector_int_) Size(ctx context.Context) (uint32, error) {
	res, err := class.CallMethod(ctx, class, "size")
	if err != nil {
		return uint32(0), err
	}

	return res.(uint32), nil
}

func NewClassVector_int_(e embind.Engine, ctx context.Context) (*ClassVector_int_, error) {
	res, err := e.CallPublicSymbol(ctx, "vector_int_")
	if err != nil {
		return nil, err
	}

	return res.(*ClassVector_int_), nil
}
