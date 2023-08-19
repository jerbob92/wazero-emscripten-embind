package embind

import "fmt"

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
