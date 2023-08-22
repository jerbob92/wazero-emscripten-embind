package main

import (
	"context"
	_ "embed"
	"log"
	"os"

	"github.com/jerbob92/wazero-emscripten-embind"
	"github.com/jerbob92/wazero-emscripten-embind/examples/classes/generated"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/emscripten"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed wasm/classes.wasm
var wasm []byte

func main() {
	ctx := context.Background()
	runtimeConfig := wazero.NewRuntimeConfig()
	r := wazero.NewRuntimeWithConfig(ctx, runtimeConfig)
	defer r.Close(ctx)

	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		log.Fatal(err)
	}

	compiledModule, err := r.CompileModule(ctx, wasm)
	if err != nil {
		log.Fatal(err)
	}

	builder := r.NewHostModuleBuilder("env")

	emscriptenExporter, err := emscripten.NewFunctionExporterForModule(compiledModule)
	if err != nil {
		log.Fatal(err)
	}

	emscriptenExporter.ExportFunctions(builder)

	engine := embind.CreateEngine(embind.NewConfig())

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
		WithStartFunctions("_initialize").
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithName("")

	ctx = engine.Attach(ctx)
	_, err = r.InstantiateModule(ctx, compiledModule, moduleConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = generated.Attach(engine)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new class.
	newClassInstance, err := generated.NewClassMyClass(engine, ctx, 23, "test123")
	if err != nil {
		log.Fatal(err)
	}

	err = generated.PrintMyClass(engine, ctx, newClassInstance)
	if err != nil {
		log.Fatal(err)
	}

	err = newClassInstance.SetX(ctx, 42)
	if err != nil {
		log.Fatal(err)
	}

	err = generated.PrintMyClass(engine, ctx, newClassInstance)
	if err != nil {
		log.Fatal(err)
	}

	err = newClassInstance.IncrementX(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = generated.PrintMyClass(engine, ctx, newClassInstance)
	if err != nil {
		log.Fatal(err)
	}

	yValue, err := generated.ClassMyClassStaticGetStringFromInstance(engine, ctx, newClassInstance)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("y from static: %s", yValue)
}
