package embind

import (
	"fmt"
	internal "github.com/jerbob92/wazero-emscripten-embind/internal"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type wazeroEngine struct {
	internal.IEngine
	config internal.IEngineConfig
}

func (we *wazeroEngine) NewFunctionExporterForModule(guest wazero.CompiledModule) FunctionExporter {
	return &functionExporter{
		config: we.config,
		guest:  guest,
	}
}

// FunctionExporter configures the functions in the "env" module used by
// Emscripten embind.
type FunctionExporter interface {
	// ExportFunctions builds functions to export with a wazero.HostModuleBuilder
	// named "env".
	ExportFunctions(wazero.HostModuleBuilder) error
}

type functionExporter struct {
	config internal.IEngineConfig
	guest  wazero.CompiledModule
}

type unexportedFunctionError struct {
	name string
}

func (e unexportedFunctionError) Error() string {
	return fmt.Sprintf("you need to export the \"%s\" function to make embind work, you can do this using the \"EXPORTED_FUNCTIONS\" option in Emscripten during compilation, you will need to prepend exports with an underscore, so you have to add \"_%s\" to the list", e.name, e.name)
}

// ExportFunctions implements FunctionExporter.ExportFunctions
func (e functionExporter) ExportFunctions(b wazero.HostModuleBuilder) error {
	// First validate whether required functions are available.
	requiredFunctions := []string{"free", "malloc", "__getTypeName"}
	exportedFunctions := e.guest.ExportedFunctions()
	for i := range requiredFunctions {
		requiredFunction := requiredFunctions[i]
		if _, ok := exportedFunctions[requiredFunction]; !ok {
			return unexportedFunctionError{
				name: requiredFunction,
			}
		}
	}

	b.NewFunctionBuilder().
		WithName("_embind_register_function").
		WithParameterNames("name", "argCount", "rawArgTypesAddr", "signature", "rawInvoker", "fn", "isAsync").
		WithGoModuleFunction(internal.RegisterFunction, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_function")

	b.NewFunctionBuilder().
		WithName("_embind_register_void").
		WithParameterNames("rawType", "name").
		WithGoModuleFunction(internal.RegisterVoid, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_void")

	b.NewFunctionBuilder().
		WithName("_embind_register_bool").
		WithParameterNames("rawType", "name", "size", "trueValue", "falseValue").
		WithGoModuleFunction(internal.RegisterBool, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_bool")

	b.NewFunctionBuilder().
		WithName("_embind_register_integer").
		WithParameterNames("rawType", "name", "size", "minRange", "maxRange").
		WithGoModuleFunction(internal.RegisterInteger, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_integer")

	b.NewFunctionBuilder().
		WithName("_embind_register_bigint").
		WithParameterNames("rawType", "name", "size", "minRange", "maxRange").
		WithGoModuleFunction(internal.RegisterBigInt, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI64, api.ValueTypeI64}, []api.ValueType{}).
		Export("_embind_register_bigint")

	b.NewFunctionBuilder().
		WithName("_embind_register_float").
		WithParameterNames("rawType", "name", "size").
		WithGoModuleFunction(internal.RegisterFloat, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_float")

	b.NewFunctionBuilder().
		WithName("_embind_register_std_string").
		WithParameterNames("rawType", "name").
		WithGoModuleFunction(internal.RegisterStdString, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_std_string")

	b.NewFunctionBuilder().
		WithName("_embind_register_std_wstring").
		WithParameterNames("rawType", "charSize", "name").
		WithGoModuleFunction(internal.RegisterStdWString, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_std_wstring")

	b.NewFunctionBuilder().
		WithName("_embind_register_memory_view").
		WithParameterNames("rawType", "dataTypeIndex", "name").
		WithGoModuleFunction(internal.RegisterMemoryView, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_memory_view")

	b.NewFunctionBuilder().
		WithName("_embind_register_emval").
		WithParameterNames("rawType", "name").
		WithGoModuleFunction(internal.RegisterEmval, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_emval")

	b.NewFunctionBuilder().
		WithName("_embind_register_constant").
		WithParameterNames("rawType", "type", "value").
		WithGoModuleFunction(internal.Constant, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeF64}, []api.ValueType{}).
		Export("_embind_register_constant")

	b.NewFunctionBuilder().
		WithName("_embind_register_enum").
		WithParameterNames("rawType", "name", "size", "isSigned").
		WithGoModuleFunction(internal.RegisterEnum, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_enum")

	b.NewFunctionBuilder().
		WithName("_embind_register_enum_value").
		WithParameterNames("rawEnumType", "name", "enumValue").
		WithGoModuleFunction(internal.RegisterEnumValue, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_enum_value")

	b.NewFunctionBuilder().
		WithName("_emval_take_value").
		WithParameterNames("type", "arg").
		WithResultNames("emval_handle").
		WithGoModuleFunction(internal.EmvalTakeValue, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export("_emval_take_value")

	b.NewFunctionBuilder().
		WithName("_emval_incref").
		WithParameterNames("handle").
		WithGoModuleFunction(internal.EmvalIncref, []api.ValueType{api.ValueTypeI32}, []api.ValueType{}).
		Export("_emval_incref")

	b.NewFunctionBuilder().
		WithName("_emval_decref").
		WithParameterNames("handle").
		WithGoModuleFunction(internal.EmvalDecref, []api.ValueType{api.ValueTypeI32}, []api.ValueType{}).
		Export("_emval_decref")

	b.NewFunctionBuilder().
		WithName("_emval_register_symbol").
		WithParameterNames("address").
		WithGoModuleFunction(internal.EmvalRegisterSymbol, []api.ValueType{api.ValueTypeI32}, []api.ValueType{}).
		Export("_emval_register_symbol")

	b.NewFunctionBuilder().
		WithName("_emval_get_global").
		WithParameterNames("name").
		WithResultNames("handle").
		WithGoModuleFunction(internal.EmvalGetGlobal, []api.ValueType{api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export("_emval_get_global")

	b.NewFunctionBuilder().
		WithName("_emval_as").
		WithParameterNames("handle", "returnType", "destructorsRef").
		WithResultNames("val").
		WithGoModuleFunction(internal.EmvalAs, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeF64}).
		Export("_emval_as")

	b.NewFunctionBuilder().
		WithName("_emval_new").
		WithParameterNames("handle", "argCount", "argTypes", "args").
		WithResultNames("val").
		WithGoModuleFunction(internal.EmvalNew, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export("_emval_new")

	b.NewFunctionBuilder().
		WithName("_emval_set_property").
		WithParameterNames("handle", "key", "value").
		WithGoModuleFunction(internal.EmvalSetProperty, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_emval_set_property")

	b.NewFunctionBuilder().
		WithName("_emval_get_property").
		WithParameterNames("handle", "key").
		WithResultNames("value").
		WithGoModuleFunction(internal.EmvalGetProperty, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export("_emval_get_property")

	b.NewFunctionBuilder().
		WithName("_emval_new_cstring").
		WithParameterNames("v").
		WithResultNames("handle").
		WithGoModuleFunction(internal.EmvalNewCString, []api.ValueType{api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export("_emval_new_cstring")

	b.NewFunctionBuilder().
		WithName("_emval_run_destructors").
		WithParameterNames("handle").
		WithGoModuleFunction(internal.EmvalRunDestructors, []api.ValueType{api.ValueTypeI32}, []api.ValueType{}).
		Export("_emval_run_destructors")

	b.NewFunctionBuilder().
		WithName("_emval_get_method_caller").
		WithParameterNames("argCount", "argTypes").
		WithResultNames("id").
		WithGoModuleFunction(internal.EmvalGetMethodCaller, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export("_emval_get_method_caller")

	b.NewFunctionBuilder().
		WithName("_emval_call_method").
		WithParameterNames("caller", "id", "methodName", "destructorsRef", "args").
		WithResultNames("value").
		WithGoModuleFunction(internal.EmvalCallMethod, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeF64}).
		Export("_emval_call_method")

	b.NewFunctionBuilder().
		WithName("_emval_call_void_method").
		WithParameterNames("caller", "id", "methodName", "args").
		WithGoModuleFunction(internal.EmvalCallVoidMethod, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_emval_call_void_method")

	b.NewFunctionBuilder().
		WithName("_embind_register_class").
		WithParameterNames("rawType", "rawPointerType", "rawConstPointerType", "baseClassRawType", "getActualTypeSignature", "getActualType", "upcastSignature", "upcast", "downcastSignature", "downcast", "name", "destructorSignature", "rawDestructor").
		WithGoModuleFunction(internal.RegisterClass, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_class")

	b.NewFunctionBuilder().
		WithName("_embind_register_class_constructor").
		WithParameterNames("rawClassType", "argCount", "rawArgTypesAddr", "invokerSignature", "invoker", "rawConstructor").
		WithGoModuleFunction(internal.RegisterClassConstructor, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_class_constructor")

	b.NewFunctionBuilder().
		WithName("_embind_register_class_function").
		WithParameterNames(
			"rawClassType",
			"methodName",
			"argCount",
			"rawArgTypesAddr", // [ReturnType, ThisType, Args...]
			"invokerSignature",
			"rawInvoker",
			"context",
			"isPureVirtual",
			"isAsync",
		).
		WithGoModuleFunction(internal.RegisterClassFunction, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_class_function")

	b.NewFunctionBuilder().
		WithName("_embind_register_class_class_function").
		WithParameterNames(
			"rawClassType",
			"methodName",
			"argCount",
			"rawArgTypesAddr", // [ReturnType, ThisType, Args...]
			"invokerSignature",
			"rawInvoker",
			"fn",
			"isAsync",
		).
		WithGoModuleFunction(internal.RegisterClassClassFunction, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_class_class_function")

	b.NewFunctionBuilder().
		WithName("_embind_register_class_property").
		WithParameterNames(
			"classType",
			"fieldName",
			"getterReturnType",
			"getterSignature",
			"getter",
			"getterContext",
			"setterArgumentType",
			"setterSignature",
			"setter",
			"setterContext",
		).
		WithGoModuleFunction(internal.RegisterClassProperty, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_class_property")

	b.NewFunctionBuilder().
		WithName("_embind_register_value_array").
		WithParameterNames(
			"rawType",
			"name",
			"constructorSignature",
			"rawConstructor",
			"destructorSignature",
			"rawDestructor",
		).
		WithGoModuleFunction(internal.RegisterValueArray, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_value_array")

	b.NewFunctionBuilder().
		WithName("_embind_register_value_array_element").
		WithParameterNames(
			"rawTupleType",
			"getterReturnType",
			"getterSignature",
			"getter",
			"getterContext",
			"setterArgumentType",
			"setterSignature",
			"setter",
			"setterContext",
		).
		WithGoModuleFunction(internal.RegisterValueArrayElement, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_value_array_element")

	b.NewFunctionBuilder().
		WithName("_embind_finalize_value_array").
		WithParameterNames("rawTupleType").
		WithGoModuleFunction(internal.FinalizeValueArray, []api.ValueType{api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_finalize_value_array")

	b.NewFunctionBuilder().
		WithName("_embind_register_value_object").
		WithParameterNames(
			"rawType",
			"name",
			"constructorSignature",
			"rawConstructor",
			"destructorSignature",
			"rawDestructor",
		).
		WithGoModuleFunction(internal.RegisterValueObject, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_value_object")

	b.NewFunctionBuilder().
		WithName("_embind_register_value_object_field").
		WithParameterNames(
			"structType",
			"fieldName",
			"getterReturnType",
			"getterSignature",
			"getter",
			"getterContext",
			"setterArgumentType",
			"setterSignature",
			"setter",
			"setterContext",
		).
		WithGoModuleFunction(internal.RegisterValueObjectField, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_register_value_object_field")

	b.NewFunctionBuilder().
		WithName("_embind_finalize_value_object").
		WithParameterNames("structType").
		WithGoModuleFunction(internal.FinalizeValueObject, []api.ValueType{api.ValueTypeI32}, []api.ValueType{}).
		Export("_embind_finalize_value_object")

	return nil
}
