// Code generated by wazero-emscripten-embind, DO NOT EDIT.
package generated

import (
	"context"

	"github.com/jerbob92/wazero-emscripten-embind"
)

func Enum_in_enum_out(e embind.Engine, ctx context.Context, arg0 EnumNewStyle) (EnumOldStyle, error) {
	res, err := e.CallPublicSymbol(ctx, "enum_in_enum_out", arg0)
	if err != nil {
		return EnumOldStyle(0), err
	}
	if res == nil {
		return EnumOldStyle(0), nil
	}
	return res.(EnumOldStyle), nil
}
