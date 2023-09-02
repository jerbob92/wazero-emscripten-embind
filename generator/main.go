package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jerbob92/wazero-emscripten-embind/generator/generator"
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

	if wasm == nil {
		log.Fatal("No wasm file given")
	}

	wasmData, err := os.ReadFile(*wasm)
	if err != nil {
		log.Fatal(err)
	}

	err = generator.Generate(dir, fileName, wasmData, *initFunction)
	if err != nil {
		log.Fatal(err)
	}
}
