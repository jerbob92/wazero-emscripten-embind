package embind

import (
	"fmt"
)

type IConstant interface {
	Name() string
	Value() any
	Type() IType
}

type registeredConstant struct {
	registeredType registeredType
	name           string
	hasCppValue    bool
	cppValue       any
	rawCppValue    uint64
	hasGoValue     bool
	goValue        any
}

func (rc *registeredConstant) validate() error {
	if rc.hasGoValue && rc.hasCppValue {
		if rc.goValue != rc.cppValue {
			return fmt.Errorf("constant %s has a different value in Go than in C++ (go: %v (%T), cpp: %v (%T))", rc.name, rc.goValue, rc.goValue, rc.cppValue, rc.cppValue)
		}
	}

	return nil
}

func (rc *registeredConstant) Name() string {
	return rc.name
}

func (rc *registeredConstant) Value() any {
	return rc.cppValue
}

func (rc *registeredConstant) Type() IType {
	return &exposedType{registeredType: rc.registeredType}
}

func (e *engine) GetConstants() []IConstant {
	constants := make([]IConstant, 0)
	for i := range e.registeredConstants {
		constants = append(constants, e.registeredConstants[i])
	}
	return constants
}
