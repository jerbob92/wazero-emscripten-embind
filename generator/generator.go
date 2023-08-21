package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"github.com/jerbob92/wazero-emscripten-embind"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/emscripten"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"go/token"
	"golang.org/x/tools/go/packages"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"unicode"
)

var (
	//go:embed templates/*
	templates embed.FS
)
var (
	packagePath  string
	fileName     string
	initFunction *string
	wasm         *string
	verbose      *bool
)

func init() {
	fileName = os.Getenv("GOFILE")
	wasm = flag.String("wasm", "", "the wasm file to process")
	initFunction = flag.String("init", "_initialize", "the function to execute to make Emscripten register the types")
	verbose = flag.Bool("v", false, "enable verbose logging")
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of wazero-emscripten-embind/generator:\n")
	fmt.Fprintf(os.Stderr, "TODO\n")
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	dir, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()
	pkgs, err := packages.Load(&packages.Config{
		Fset: fset,
		Mode: packages.NeedSyntax | packages.NeedName | packages.NeedModule | packages.NeedTypes | packages.NeedTypesInfo,
	}, fmt.Sprintf("file=%s", fileName))
	if err != nil {
		panic(err)
	}

	if wasm == nil {
		log.Fatal("No wasm file given")
	}

	ctx := context.Background()
	runtimeConfig := wazero.NewRuntimeConfigInterpreter()
	r := wazero.NewRuntimeWithConfig(ctx, runtimeConfig)
	defer r.Close(ctx)

	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		log.Fatal(err)
	}

	wasmData, err := os.ReadFile(*wasm)
	if err != nil {
		log.Fatal(err)
	}

	compiledModule, err := r.CompileModule(ctx, wasmData)
	if err != nil {
		log.Fatal(err)
	}

	builder := r.NewHostModuleBuilder("env")
	emscriptenExporter, err := emscripten.NewFunctionExporterForModule(compiledModule)
	if err != nil {
		log.Fatal(err)
	}
	emscriptenExporter.ExportFunctions(builder)

	engine := embind.CreateEngine(nil)
	ctx = engine.Attach(ctx)

	embindExporter := engine.NewFunctionExporterForModule(compiledModule)
	err = embindExporter.ExportFunctions(builder)
	if err != nil {
		log.Fatal(err)
	}

	_, err = builder.Instantiate(ctx)
	if err != nil {
		log.Fatal(err)
	}

	moduleConfig := wazero.NewModuleConfig().
		WithStartFunctions("").
		WithName("")

	mod, err := r.InstantiateModule(ctx, compiledModule, moduleConfig)
	if err != nil {
		log.Fatal(err)
	}

	initFunc := mod.ExportedFunction(*initFunction)
	if initFunc == nil {
		log.Fatalf("init function %s does not exist", *initFunction)
	}

	res, err := initFunc.Call(ctx)
	if res != nil {
		log.Fatal(fmt.Errorf("could not call init function %w", err))
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
		if isClass || strings.HasPrefix(convertedName, "[]") || strings.HasPrefix(convertedName, "map[") {
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

		symbol := TemplateSymbol{
			Symbol:        symbols[i].Symbol(),
			GoName:        "Symbol" + generateGoName(symbols[i].Symbol()),
			ArgumentTypes: argumentTypes,
			ReturnType:    typeNameToGeneratedName(symbols[i].ReturnType().Type(), symbols[i].ReturnType().IsClass(), symbols[i].ReturnType().IsEnum()),
			ErrorValue:    typeNameToErrorValue(symbols[i].ReturnType().Type(), symbols[i].ReturnType().IsClass(), symbols[i].ReturnType().IsEnum()),
		}

		data.Symbols = append(data.Symbols, symbol)
	}

	sort.Slice(data.Symbols, func(i, j int) bool {
		return data.Symbols[i].GoName < data.Symbols[j].GoName
	})

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

		properties := classes[i].Properties()
		for pi := range properties {
			property := TemplateClassProperty{
				Name:       properties[pi].Name(),
				GoName:     generateGoName(properties[pi].Name()),
				ReadOnly:   properties[pi].ReadOnly(),
				GetterType: typeNameToGeneratedName(properties[pi].GetterType().Type(), properties[pi].GetterType().IsClass(), properties[pi].GetterType().IsEnum()),
				ErrorValue: typeNameToErrorValue(properties[pi].GetterType().Type(), properties[pi].GetterType().IsClass(), properties[pi].GetterType().IsEnum()),
			}

			if !property.ReadOnly {
				property.SetterType = typeNameToGeneratedName(properties[pi].SetterType().Type(), properties[pi].SetterType().IsClass(), properties[pi].SetterType().IsEnum())
			}

			class.Properties = append(class.Properties, property)
		}

		data.Classes = append(data.Classes, class)
	}

	sort.Slice(data.Classes, func(i, j int) bool {
		return data.Classes[i].GoName < data.Classes[j].GoName
	})

	if len(data.Classes) > 0 {
		ExecuteTemplate(templates, "classes.tmpl", path.Join(dir, "classes.go"), data)
	} else {
		_ = os.Remove(path.Join(dir, "classes.go"))
	}
	if len(data.Constants) > 0 {
		ExecuteTemplate(templates, "constants.tmpl", path.Join(dir, "constants.go"), data)
	} else {
		_ = os.Remove(path.Join(dir, "constants.go"))
	}
	if len(data.Symbols) > 0 {
		ExecuteTemplate(templates, "symbols.tmpl", path.Join(dir, "symbols.go"), data)
	} else {
		_ = os.Remove(path.Join(dir, "symbols.go"))
	}
	if len(data.Enums) > 0 {
		ExecuteTemplate(templates, "enums.tmpl", path.Join(dir, "enums.go"), data)
	} else {
		_ = os.Remove(path.Join(dir, "enums.go"))
	}
	ExecuteTemplate(templates, "engine.tmpl", path.Join(dir, "engine.go"), data)
}

var TemplateFunctions = template.FuncMap{
	"lower": strings.ToLower,
}

func ExecuteTemplate(tmpl *template.Template, name string, path string, data TemplateData) {
	writer, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer writer.Close()

	err = tmpl.ExecuteTemplate(writer, name, data)
	if err != nil {
		panic(err)
	}
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
	Name         string
	GoName       string
	Constructors []TemplateClassConstructor
	Properties   []TemplateClassProperty
	Methods      []TemplateClassMethods
}

type TemplateClassProperty struct {
	GoName     string
	Name       string
	SetterType string
	GetterType string
	ReadOnly   bool
	ErrorValue string
}

type TemplateClassMethods struct {
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
