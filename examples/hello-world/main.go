package main

import (
	"context"
	_ "embed"
	"log"
	"os"

	"github.com/jerbob92/wazero-emscripten-embind"
	"github.com/jerbob92/wazero-emscripten-embind/examples/hello-world/generated"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/emscripten"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed wasm/hello_world.wasm
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

	err = generated.Hello_world(engine, ctx, "Wazero user")
	if err != nil {
		log.Fatal(err)
	}
}
