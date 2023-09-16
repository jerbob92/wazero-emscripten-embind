package generator

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"go/format"
	"go/token"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/jerbob92/wazero-emscripten-embind"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/emscripten"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"golang.org/x/tools/go/packages"
)

var (
	//go:embed templates/*
	templates embed.FS
)

func Generate(dir string, fileName string, wasm []byte, initFunction string) error {
	fset := token.NewFileSet()
	pkgs, err := packages.Load(&packages.Config{
		Fset: fset,
		Mode: packages.NeedSyntax | packages.NeedName | packages.NeedModule | packages.NeedTypes | packages.NeedTypesInfo,
	}, fmt.Sprintf("file=%s", fileName))
	if err != nil {
		return err
	}

	ctx := context.Background()
	runtimeConfig := wazero.NewRuntimeConfigInterpreter()
	r := wazero.NewRuntimeWithConfig(ctx, runtimeConfig)
	defer r.Close(ctx)

	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		return err
	}

	compiledModule, err := r.CompileModule(ctx, wasm)
	if err != nil {
		return err
	}

	builder := r.NewHostModuleBuilder("env")
	emscriptenExporter, err := emscripten.NewFunctionExporterForModule(compiledModule)
	if err != nil {
		return err
	}
	emscriptenExporter.ExportFunctions(builder)

	engine := embind.CreateEngine(nil)
	ctx = engine.Attach(ctx)

	embindExporter := engine.NewFunctionExporterForModule(compiledModule)
	err = embindExporter.ExportFunctions(builder)
	if err != nil {
		return err
	}

	_, err = builder.Instantiate(ctx)
	if err != nil {
		return err
	}

	moduleConfig := wazero.NewModuleConfig().
		WithStartFunctions("").
		WithName("")

	mod, err := r.InstantiateModule(ctx, compiledModule, moduleConfig)
	if err != nil {
		return err
	}

	initFunc := mod.ExportedFunction(initFunction)
	if initFunc == nil {
		log.Fatalf("init function %s does not exist", initFunction)
	}

	res, err := initFunc.Call(ctx)
	if res != nil {
		return fmt.Errorf("could not call init function %w", err)
	}

	packageName := pkgs[0].Name
	packagePath := pkgs[0].PkgPath

	templates, err := template.New("").
		Funcs(TemplateFunctions). // Custom functions
		ParseFS(templates, "templates/*.tmpl")
	if err != nil {
		panic(err)
	}

	data := TemplateData{
		Pkg:       packageName,
		PkgPath:   packagePath,
		Symbols:   []TemplateSymbol{},
		Constants: []TemplateConstant{},
		Enums:     []TemplateEnum{},
		Classes:   []TemplateClass{},
	}

	generateGoName := func(name string) string {
		if len(name) == 0 {
			return name
		}
		upperFirst := string(unicode.ToUpper(rune(name[0]))) + name[1:]
		return upperFirst
	}

	typeNameToGeneratedName := func(name string, isClass, isEnum bool) string {
		if isClass {
			name = strings.TrimPrefix(name, "*")
			name = "Class" + generateGoName(name)
			name = "*" + name
		} else if isEnum {
			name = "Enum" + generateGoName(name)
		}

		return name
	}

	typeNameToErrorValue := func(name string, isClass, isEnum bool) string {
		convertedName := typeNameToGeneratedName(name, isClass, isEnum)
		if isClass || convertedName == "any" || strings.HasPrefix(convertedName, "[]") || strings.HasPrefix(convertedName, "map[") {
			return "nil"
		}

		if convertedName == "string" {
			return "\"\""
		}

		if convertedName == "bool" {
			return convertedName + "(false)"
		}

		// Default other types take 0
		return convertedName + "(0)"
	}

	constants := engine.GetConstants()
	for i := range constants {
		constant := TemplateConstant{
			Name:        constants[i].Name(),
			GoName:      "Constant_" + constants[i].Name(),
			Value:       fmt.Sprintf("%v", constants[i].Value()),
			GoType:      typeNameToGeneratedName(constants[i].Type().Type(), constants[i].Type().IsClass(), constants[i].Type().IsEnum()),
			ValuePrefix: "(",
			ValueSuffix: ")",
		}

		if constant.GoType == "string" {
			constant.Value = "\"" + constant.Value + "\""
		}

		data.Constants = append(data.Constants, constant)
	}

	sort.Slice(data.Constants, func(i, j int) bool {
		return data.Constants[i].Name < data.Constants[j].Name
	})

	symbols := engine.GetSymbols()
	for i := range symbols {
		exposedArgumentTypes := symbols[i].ArgumentTypes()
		argumentTypes := make([]string, len(exposedArgumentTypes))
		for i := range exposedArgumentTypes {
			argumentTypes[i] = typeNameToGeneratedName(exposedArgumentTypes[i].Type(), exposedArgumentTypes[i].IsClass(), exposedArgumentTypes[i].IsEnum())
		}

		goName := generateGoName(symbols[i].Symbol())
		if symbols[i].IsOverload() {
			goName += strconv.Itoa(len(argumentTypes))
		}

		returnType := symbols[i].ReturnType()
		if returnType == nil {
			continue
		}

		symbol := TemplateSymbol{
			Symbol:        symbols[i].Symbol(),
			GoName:        goName,
			ArgumentTypes: argumentTypes,
			ReturnType:    typeNameToGeneratedName(returnType.Type(), returnType.IsClass(), returnType.IsEnum()),
			ErrorValue:    typeNameToErrorValue(returnType.Type(), returnType.IsClass(), returnType.IsEnum()),
		}

		data.Symbols = append(data.Symbols, symbol)
	}

	sort.Slice(data.Symbols, func(i, j int) bool {
		if data.Symbols[i].GoName == data.Symbols[j].GoName {
			return data.Symbols[i].Symbol < data.Symbols[j].Symbol
		}
		return data.Symbols[i].GoName < data.Symbols[j].GoName
	})

	// Prevent duplicate names.
	seenNames := map[string]bool{}
	for i := range data.Symbols {
		_, ok := seenNames[data.Symbols[i].GoName]
		if ok {
			data.Symbols[i].GoName += "_"
		}

		seenNames[data.Symbols[i].GoName] = true
	}

	enums := engine.GetEnums()
	for i := range enums {
		enum := TemplateEnum{
			Name:   enums[i].Name(),
			GoName: "Enum" + generateGoName(enums[i].Name()),
			GoType: typeNameToGeneratedName(enums[i].Type().Type(), enums[i].Type().IsClass(), enums[i].Type().IsEnum()),
			Values: []TemplateEnumValue{},
		}

		values := enums[i].Values()
		for i := range values {
			enum.Values = append(enum.Values, TemplateEnumValue{
				Name:   values[i].Name(),
				GoName: values[i].Name(),
				Value:  fmt.Sprintf("%v", values[i].Value()),
			})
		}

		sort.Slice(enum.Values, func(i, j int) bool {
			return enum.Values[i].GoName < enum.Values[j].GoName
		})

		data.Enums = append(data.Enums, enum)
	}

	sort.Slice(data.Enums, func(i, j int) bool {
		return data.Enums[i].GoName < data.Enums[j].GoName
	})

	classes := engine.GetClasses()
	for i := range classes {
		class := TemplateClass{
			Name:         classes[i].Name(),
			GoName:       "Class" + generateGoName(classes[i].Name()),
			Constructors: []TemplateClassConstructor{},
		}

		constructors := classes[i].Constructors()
		for ci := range constructors {
			exposedArgumentTypes := constructors[ci].ArgumentTypes()
			argumentTypes := make([]string, len(exposedArgumentTypes))
			for i := range exposedArgumentTypes {
				argumentTypes[i] = typeNameToGeneratedName(exposedArgumentTypes[i].Type(), exposedArgumentTypes[i].IsClass(), exposedArgumentTypes[i].IsEnum())
			}

			constructor := TemplateClassConstructor{
				Name:          constructors[ci].Name(),
				ArgumentTypes: argumentTypes,
			}

			class.Constructors = append(class.Constructors, constructor)
		}

		sort.Slice(class.Constructors, func(i, j int) bool {
			return class.Constructors[i].Name < class.Constructors[j].Name
		})

		properties := classes[i].Properties()
		for pi := range properties {
			getterType := properties[pi].GetterType()
			if getterType == nil {
				continue
			}

			property := TemplateClassProperty{
				Name:       properties[pi].Name(),
				GoName:     generateGoName(properties[pi].Name()),
				ReadOnly:   properties[pi].ReadOnly(),
				GetterType: typeNameToGeneratedName(getterType.Type(), getterType.IsClass(), getterType.IsEnum()),
				ErrorValue: typeNameToErrorValue(getterType.Type(), getterType.IsClass(), getterType.IsEnum()),
			}

			if !property.ReadOnly {
				setterType := properties[pi].SetterType()
				if setterType == nil {
					continue
				}
				property.SetterType = typeNameToGeneratedName(setterType.Type(), setterType.IsClass(), setterType.IsEnum())
			}

			class.Properties = append(class.Properties, property)
		}

		sort.Slice(class.Properties, func(i, j int) bool {
			return class.Properties[i].GoName < class.Properties[j].GoName
		})

		staticProperties := classes[i].StaticProperties()
		for pi := range staticProperties {
			getterType := staticProperties[pi].GetterType()
			if getterType == nil {
				continue
			}

			property := TemplateClassProperty{
				Name:       staticProperties[pi].Name(),
				GoName:     generateGoName(staticProperties[pi].Name()),
				ReadOnly:   staticProperties[pi].ReadOnly(),
				GetterType: typeNameToGeneratedName(getterType.Type(), getterType.IsClass(), getterType.IsEnum()),
				ErrorValue: typeNameToErrorValue(getterType.Type(), getterType.IsClass(), getterType.IsEnum()),
			}

			if !property.ReadOnly {
				setterType := staticProperties[pi].SetterType()
				if setterType == nil {
					continue
				}
				property.SetterType = typeNameToGeneratedName(setterType.Type(), setterType.IsClass(), setterType.IsEnum())
			}

			class.StaticProperties = append(class.StaticProperties, property)
		}

		sort.Slice(class.StaticProperties, func(i, j int) bool {
			return class.StaticProperties[i].GoName < class.StaticProperties[j].GoName
		})

		methods := classes[i].Methods()
		for mi := range methods {
			exposedArgumentTypes := methods[mi].ArgumentTypes()
			argumentTypes := make([]string, len(exposedArgumentTypes))
			for i := range exposedArgumentTypes {
				argumentTypes[i] = typeNameToGeneratedName(exposedArgumentTypes[i].Type(), exposedArgumentTypes[i].IsClass(), exposedArgumentTypes[i].IsEnum())
			}

			goName := generateGoName(methods[mi].Symbol())
			if methods[mi].IsOverload() {
				goName += strconv.Itoa(len(argumentTypes))
			}

			returnType := methods[mi].ReturnType()
			if returnType == nil {
				continue
			}

			method := TemplateClassMethod{
				Name:          methods[mi].Symbol(),
				GoName:        goName,
				ArgumentTypes: argumentTypes,
				ReturnType:    typeNameToGeneratedName(returnType.Type(), returnType.IsClass(), returnType.IsEnum()),
				ErrorValue:    typeNameToErrorValue(returnType.Type(), returnType.IsClass(), returnType.IsEnum()),
			}

			class.Methods = append(class.Methods, method)
		}

		sort.Slice(class.Methods, func(i, j int) bool {
			return class.Methods[i].GoName < class.Methods[j].GoName
		})

		staticMethods := classes[i].StaticMethods()
		for smi := range staticMethods {
			exposedArgumentTypes := staticMethods[smi].ArgumentTypes()
			argumentTypes := make([]string, len(exposedArgumentTypes))
			for i := range exposedArgumentTypes {
				argumentTypes[i] = typeNameToGeneratedName(exposedArgumentTypes[i].Type(), exposedArgumentTypes[i].IsClass(), exposedArgumentTypes[i].IsEnum())
			}

			goName := generateGoName(staticMethods[smi].Symbol())
			if staticMethods[smi].IsOverload() {
				goName += strconv.Itoa(len(argumentTypes))
			}

			returnType := staticMethods[smi].ReturnType()
			if returnType == nil {
				continue
			}

			method := TemplateClassMethod{
				Name:          staticMethods[smi].Symbol(),
				GoName:        goName,
				ArgumentTypes: argumentTypes,
				ReturnType:    typeNameToGeneratedName(returnType.Type(), returnType.IsClass(), returnType.IsEnum()),
				ErrorValue:    typeNameToErrorValue(returnType.Type(), returnType.IsClass(), returnType.IsEnum()),
			}

			class.StaticMethods = append(class.StaticMethods, method)
		}

		sort.Slice(class.StaticMethods, func(i, j int) bool {
			return class.StaticMethods[i].GoName < class.StaticMethods[j].GoName
		})

		data.Classes = append(data.Classes, class)
	}

	sort.Slice(data.Classes, func(i, j int) bool {
		return data.Classes[i].GoName < data.Classes[j].GoName
	})

	if len(data.Classes) > 0 {
		err = ExecuteTemplate(templates, "classes.tmpl", path.Join(dir, "classes.go"), data)
		if err != nil {
			return err
		}
	} else {
		_ = os.Remove(path.Join(dir, "classes.go"))
	}
	if len(data.Constants) > 0 {
		err = ExecuteTemplate(templates, "constants.tmpl", path.Join(dir, "constants.go"), data)
		if err != nil {
			return err
		}
	} else {
		_ = os.Remove(path.Join(dir, "constants.go"))
	}
	if len(data.Symbols) > 0 {
		err = ExecuteTemplate(templates, "functions.tmpl", path.Join(dir, "functions.go"), data)
		if err != nil {
			return err
		}
	} else {
		_ = os.Remove(path.Join(dir, "functions.go"))
	}
	if len(data.Enums) > 0 {
		err = ExecuteTemplate(templates, "enums.tmpl", path.Join(dir, "enums.go"), data)
		if err != nil {
			return err
		}
	} else {
		_ = os.Remove(path.Join(dir, "enums.go"))
	}

	err = ExecuteTemplate(templates, "engine.tmpl", path.Join(dir, "engine.go"), data)
	if err != nil {
		return err
	}

	return nil
}

var TemplateFunctions = template.FuncMap{
	"lower": strings.ToLower,
}

func ExecuteTemplate(tmpl *template.Template, name string, path string, data TemplateData) error {
	writer := bytes.NewBuffer(nil)
	err := tmpl.ExecuteTemplate(writer, name, data)
	if err != nil {
		return err
	}

	fileBytes := writer.Bytes()
	formattedSource, err := format.Source(fileBytes)
	if err != nil {
		return fmt.Errorf("could not format %s: %w", name, err)
	}

	fileWriter, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fileWriter.Close()
	_, err = fileWriter.Write(formattedSource)
	if err != nil {
		return err
	}
	return nil
}

type TemplateData struct {
	Pkg       string
	PkgPath   string
	Enums     []TemplateEnum
	Symbols   []TemplateSymbol
	Constants []TemplateConstant
	Classes   []TemplateClass
}

type TemplateConstant struct {
	Name        string
	GoName      string
	Value       string
	GoType      string
	ValuePrefix string
	ValueSuffix string
}

type TemplateEnum struct {
	Name   string
	GoName string
	GoType string
	Values []TemplateEnumValue
}

type TemplateEnumValue struct {
	Name   string
	GoName string
	Value  string
}

type TemplateClass struct {
	Name             string
	GoName           string
	Constructors     []TemplateClassConstructor
	Properties       []TemplateClassProperty
	StaticProperties []TemplateClassProperty
	Methods          []TemplateClassMethod
	StaticMethods    []TemplateClassMethod
}

type TemplateClassProperty struct {
	GoName     string
	Name       string
	SetterType string
	GetterType string
	ReadOnly   bool
	ErrorValue string
}

type TemplateClassMethod struct {
	GoName        string
	Name          string
	ArgumentTypes []string
	ReturnType    string
	ErrorValue    string
}

type TemplateClassConstructor struct {
	Name          string
	ArgumentTypes []string
}

type TemplateSymbol struct {
	Symbol        string
	GoName        string
	ArgumentTypes []string
	ReturnType    string
	ErrorValue    string
}
