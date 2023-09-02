// Code generated by wazero-emscripten-embind, DO NOT EDIT.
package generated

import (
	"context"

	"github.com/jerbob92/wazero-emscripten-embind"
)

func Base(e embind.Engine, ctx context.Context) (*ClassBase, error) {
	res, err := e.CallPublicSymbol(ctx, "Base")
	if err != nil {
		return nil, err
	}
	return res.(*ClassBase), nil
}

func BaseWrapper(e embind.Engine, ctx context.Context) (*ClassBaseWrapper, error) {
	res, err := e.CallPublicSymbol(ctx, "BaseWrapper")
	if err != nil {
		return nil, err
	}
	return res.(*ClassBaseWrapper), nil
}

func Bool_return_bool(e embind.Engine, ctx context.Context, arg0 bool) (bool, error) {
	res, err := e.CallPublicSymbol(ctx, "bool_return_bool", arg0)
	if err != nil {
		return bool(false), err
	}
	return res.(bool), nil
}

func Bool_return_false(e embind.Engine, ctx context.Context) (bool, error) {
	res, err := e.CallPublicSymbol(ctx, "bool_return_false")
	if err != nil {
		return bool(false), err
	}
	return res.(bool), nil
}

func Bool_return_true(e embind.Engine, ctx context.Context) (bool, error) {
	res, err := e.CallPublicSymbol(ctx, "bool_return_true")
	if err != nil {
		return bool(false), err
	}
	return res.(bool), nil
}

func C(e embind.Engine, ctx context.Context) (*ClassC, error) {
	res, err := e.CallPublicSymbol(ctx, "C")
	if err != nil {
		return nil, err
	}
	return res.(*ClassC), nil
}

func Char_return_char(e embind.Engine, ctx context.Context, arg0 int8) (int8, error) {
	res, err := e.CallPublicSymbol(ctx, "char_return_char", arg0)
	if err != nil {
		return int8(0), err
	}
	return res.(int8), nil
}

func Derived(e embind.Engine, ctx context.Context) (*ClassDerived, error) {
	res, err := e.CallPublicSymbol(ctx, "Derived")
	if err != nil {
		return nil, err
	}
	return res.(*ClassDerived), nil
}

func DoEmval(e embind.Engine, ctx context.Context) (string, error) {
	res, err := e.CallPublicSymbol(ctx, "doEmval")
	if err != nil {
		return "", err
	}
	return res.(string), nil
}

func Double_return_double(e embind.Engine, ctx context.Context, arg0 float64) (float64, error) {
	res, err := e.CallPublicSymbol(ctx, "double_return_double", arg0)
	if err != nil {
		return float64(0), err
	}
	return res.(float64), nil
}

func Enum_in_enum_out(e embind.Engine, ctx context.Context, arg0 EnumNewStyle) (EnumOldStyle, error) {
	res, err := e.CallPublicSymbol(ctx, "enum_in_enum_out", arg0)
	if err != nil {
		return EnumOldStyle(0), err
	}
	return res.(EnumOldStyle), nil
}

func FindPersonAtLocation(e embind.Engine, ctx context.Context, arg0 []any) (map[string]any, error) {
	res, err := e.CallPublicSymbol(ctx, "findPersonAtLocation", arg0)
	if err != nil {
		return nil, err
	}
	return res.(map[string]any), nil
}

func Float_return_float(e embind.Engine, ctx context.Context, arg0 float32) (float32, error) {
	res, err := e.CallPublicSymbol(ctx, "float_return_float", arg0)
	if err != nil {
		return float32(0), err
	}
	return res.(float32), nil
}

func Float_return_void(e embind.Engine, ctx context.Context, arg0 float32) error {
	_, err := e.CallPublicSymbol(ctx, "float_return_void", arg0)
	return err
}

func Function_overload0(e embind.Engine, ctx context.Context) (int32, error) {
	res, err := e.CallPublicSymbol(ctx, "function_overload")
	if err != nil {
		return int32(0), err
	}
	return res.(int32), nil
}

func Function_overload1(e embind.Engine, ctx context.Context, arg0 int32) (int32, error) {
	res, err := e.CallPublicSymbol(ctx, "function_overload", arg0)
	if err != nil {
		return int32(0), err
	}
	return res.(int32), nil
}

func GetDerivedInstance(e embind.Engine, ctx context.Context) (*ClassBase, error) {
	res, err := e.CallPublicSymbol(ctx, "getDerivedInstance")
	if err != nil {
		return nil, err
	}
	return res.(*ClassBase), nil
}

func Get_memory_view_char(e embind.Engine, ctx context.Context) (any, error) {
	res, err := e.CallPublicSymbol(ctx, "get_memory_view_char")
	if err != nil {
		return nil, err
	}
	return res.(any), nil
}

func Get_memory_view_double(e embind.Engine, ctx context.Context) (any, error) {
	res, err := e.CallPublicSymbol(ctx, "get_memory_view_double")
	if err != nil {
		return nil, err
	}
	return res.(any), nil
}

func Get_memory_view_float(e embind.Engine, ctx context.Context) (any, error) {
	res, err := e.CallPublicSymbol(ctx, "get_memory_view_float")
	if err != nil {
		return nil, err
	}
	return res.(any), nil
}

func Get_memory_view_int(e embind.Engine, ctx context.Context) (any, error) {
	res, err := e.CallPublicSymbol(ctx, "get_memory_view_int")
	if err != nil {
		return nil, err
	}
	return res.(any), nil
}

func Get_memory_view_long(e embind.Engine, ctx context.Context) (any, error) {
	res, err := e.CallPublicSymbol(ctx, "get_memory_view_long")
	if err != nil {
		return nil, err
	}
	return res.(any), nil
}

func Get_memory_view_longlong(e embind.Engine, ctx context.Context) (any, error) {
	res, err := e.CallPublicSymbol(ctx, "get_memory_view_longlong")
	if err != nil {
		return nil, err
	}
	return res.(any), nil
}

func Get_memory_view_short(e embind.Engine, ctx context.Context) (any, error) {
	res, err := e.CallPublicSymbol(ctx, "get_memory_view_short")
	if err != nil {
		return nil, err
	}
	return res.(any), nil
}

func Get_memory_view_unsigned_char(e embind.Engine, ctx context.Context) (any, error) {
	res, err := e.CallPublicSymbol(ctx, "get_memory_view_unsigned_char")
	if err != nil {
		return nil, err
	}
	return res.(any), nil
}

func Get_memory_view_unsigned_int(e embind.Engine, ctx context.Context) (any, error) {
	res, err := e.CallPublicSymbol(ctx, "get_memory_view_unsigned_int")
	if err != nil {
		return nil, err
	}
	return res.(any), nil
}

func Get_memory_view_unsigned_long(e embind.Engine, ctx context.Context) (any, error) {
	res, err := e.CallPublicSymbol(ctx, "get_memory_view_unsigned_long")
	if err != nil {
		return nil, err
	}
	return res.(any), nil
}

func Get_memory_view_unsigned_longlong(e embind.Engine, ctx context.Context) (any, error) {
	res, err := e.CallPublicSymbol(ctx, "get_memory_view_unsigned_longlong")
	if err != nil {
		return nil, err
	}
	return res.(any), nil
}

func Get_memory_view_unsigned_short(e embind.Engine, ctx context.Context) (any, error) {
	res, err := e.CallPublicSymbol(ctx, "get_memory_view_unsigned_short")
	if err != nil {
		return nil, err
	}
	return res.(any), nil
}

func Int_return_int(e embind.Engine, ctx context.Context, arg0 int32) (int32, error) {
	res, err := e.CallPublicSymbol(ctx, "int_return_int", arg0)
	if err != nil {
		return int32(0), err
	}
	return res.(int32), nil
}

func Interface(e embind.Engine, ctx context.Context) (*ClassInterface, error) {
	res, err := e.CallPublicSymbol(ctx, "Interface")
	if err != nil {
		return nil, err
	}
	return res.(*ClassInterface), nil
}

func InterfaceWrapper(e embind.Engine, ctx context.Context) (*ClassInterfaceWrapper, error) {
	res, err := e.CallPublicSymbol(ctx, "InterfaceWrapper")
	if err != nil {
		return nil, err
	}
	return res.(*ClassInterfaceWrapper), nil
}

func Long_return_long(e embind.Engine, ctx context.Context, arg0 int32) (int32, error) {
	res, err := e.CallPublicSymbol(ctx, "long_return_long", arg0)
	if err != nil {
		return int32(0), err
	}
	return res.(int32), nil
}

func Longlong_return_longlong(e embind.Engine, ctx context.Context, arg0 int64) (int64, error) {
	res, err := e.CallPublicSymbol(ctx, "longlong_return_longlong", arg0)
	if err != nil {
		return int64(0), err
	}
	return res.(int64), nil
}

func Map_int__string_(e embind.Engine, ctx context.Context) (*ClassMap_int__string_, error) {
	res, err := e.CallPublicSymbol(ctx, "map_int__string_")
	if err != nil {
		return nil, err
	}
	return res.(*ClassMap_int__string_), nil
}

func MyClass(e embind.Engine, ctx context.Context) (*ClassMyClass, error) {
	res, err := e.CallPublicSymbol(ctx, "MyClass")
	if err != nil {
		return nil, err
	}
	return res.(*ClassMyClass), nil
}

func PassThrough(e embind.Engine, ctx context.Context, arg0 *ClassMyClass) (*ClassMyClass, error) {
	res, err := e.CallPublicSymbol(ctx, "passThrough", arg0)
	if err != nil {
		return nil, err
	}
	return res.(*ClassMyClass), nil
}

func Return_map(e embind.Engine, ctx context.Context) (*ClassMap_int__string_, error) {
	res, err := e.CallPublicSymbol(ctx, "return_map")
	if err != nil {
		return nil, err
	}
	return res.(*ClassMap_int__string_), nil
}

func Return_vector(e embind.Engine, ctx context.Context) (*ClassVector_int_, error) {
	res, err := e.CallPublicSymbol(ctx, "return_vector")
	if err != nil {
		return nil, err
	}
	return res.(*ClassVector_int_), nil
}

func SetPersonAtLocation(e embind.Engine, ctx context.Context, arg0 []any, arg1 map[string]any) error {
	_, err := e.CallPublicSymbol(ctx, "setPersonAtLocation", arg0, arg1)
	return err
}

func Short_return_short(e embind.Engine, ctx context.Context, arg0 int16) (int16, error) {
	res, err := e.CallPublicSymbol(ctx, "short_return_short", arg0)
	if err != nil {
		return int16(0), err
	}
	return res.(int16), nil
}

func Std_string_return_std_string(e embind.Engine, ctx context.Context, arg0 string) (string, error) {
	res, err := e.CallPublicSymbol(ctx, "std_string_return_std_string", arg0)
	if err != nil {
		return "", err
	}
	return res.(string), nil
}

func Std_u16string_return_std_u16string(e embind.Engine, ctx context.Context, arg0 string) (string, error) {
	res, err := e.CallPublicSymbol(ctx, "std_u16string_return_std_u16string", arg0)
	if err != nil {
		return "", err
	}
	return res.(string), nil
}

func Std_wstring_return_std_wstring(e embind.Engine, ctx context.Context, arg0 string) (string, error) {
	res, err := e.CallPublicSymbol(ctx, "std_wstring_return_std_wstring", arg0)
	if err != nil {
		return "", err
	}
	return res.(string), nil
}

func Uchar_return_uchar(e embind.Engine, ctx context.Context, arg0 uint8) (uint8, error) {
	res, err := e.CallPublicSymbol(ctx, "uchar_return_uchar", arg0)
	if err != nil {
		return uint8(0), err
	}
	return res.(uint8), nil
}

func Uint_return_uint(e embind.Engine, ctx context.Context, arg0 uint32) (uint32, error) {
	res, err := e.CallPublicSymbol(ctx, "uint_return_uint", arg0)
	if err != nil {
		return uint32(0), err
	}
	return res.(uint32), nil
}

func Ulong_return_ulong(e embind.Engine, ctx context.Context, arg0 uint32) (uint32, error) {
	res, err := e.CallPublicSymbol(ctx, "ulong_return_ulong", arg0)
	if err != nil {
		return uint32(0), err
	}
	return res.(uint32), nil
}

func Ulonglong_return_ulonglong(e embind.Engine, ctx context.Context, arg0 uint64) (uint64, error) {
	res, err := e.CallPublicSymbol(ctx, "ulonglong_return_ulonglong", arg0)
	if err != nil {
		return uint64(0), err
	}
	return res.(uint64), nil
}

func Ushort_return_ushort(e embind.Engine, ctx context.Context, arg0 uint16) (uint16, error) {
	res, err := e.CallPublicSymbol(ctx, "ushort_return_ushort", arg0)
	if err != nil {
		return uint16(0), err
	}
	return res.(uint16), nil
}

func Vector_int_(e embind.Engine, ctx context.Context) (*ClassVector_int_, error) {
	res, err := e.CallPublicSymbol(ctx, "vector_int_")
	if err != nil {
		return nil, err
	}
	return res.(*ClassVector_int_), nil
}
