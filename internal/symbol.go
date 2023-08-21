package embind

import (
	"context"
	"fmt"
)

type ISymbol interface {
	Symbol() string
	ReturnType() IType
	ArgumentTypes() []IType
}

type publicSymbol struct {
	name          string
	argCount      *int32
	overloadTable map[int32]*publicSymbol
	fn            publicSymbolFn
	className     string
	argumentTypes []registeredType
	resultType    registeredType
	isStatic      bool
}

func (ps *publicSymbol) Symbol() string {
	return ps.name
}

func (ps *publicSymbol) ReturnType() IType {
	return &exposedType{ps.resultType}
}

func (ps *publicSymbol) ArgumentTypes() []IType {
	exposedTypes := make([]IType, len(ps.argumentTypes))
	for i := range ps.argumentTypes {
		exposedTypes[i] = &exposedType{ps.argumentTypes[i]}
	}
	return exposedTypes
}

type PublicSymbol interface {
	Call(ctx context.Context, this any, arguments ...any) (any, error)
}

func (e *engine) CallPublicSymbol(ctx context.Context, name string, arguments ...any) (any, error) {
	_, ok := e.publicSymbols[name]
	if !ok {
		return nil, fmt.Errorf("could not find public symbol %s", name)
	}

	ctx = e.Attach(ctx)
	res, err := e.publicSymbols[name].fn(ctx, nil, arguments...)
	if err != nil {
		return nil, fmt.Errorf("error while calling embind function %s: %w", name, err)
	}

	return res, nil
}

func (e *engine) GetSymbols() []ISymbol {
	symbols := make([]ISymbol, 0)

	for i := range e.publicSymbols {
		if e.publicSymbols[i].overloadTable != nil {
			for argCount := range e.publicSymbols[i].overloadTable {
				symbols = append(symbols, e.publicSymbols[i].overloadTable[argCount])
			}
		} else {
			symbols = append(symbols, e.publicSymbols[i])
		}
	}

	return symbols
}