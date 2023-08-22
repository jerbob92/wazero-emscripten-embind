# wazero-emscripten-embind

[![Go Reference](https://pkg.go.dev/badge/github.com/jerbob92/wazero-emscripten-embind.svg)](https://pkg.go.dev/github.com/jerbob92/wazero-emscripten-embind)

:rocket: *Emscripten Embind support for Go using Wazero* :rocket:

## Features

* Full support for all Embind features
* Full code generator for Embind bindings:
    * Functions
    * Classes
    * Enums
    * Constants
* Typed data and function signatures where possible
* Ability to call Go code from Embind using Emval
* Communicate between guest and host without worrying about data encoding/decoding

## What does Embind do?

[Embind](https://emscripten.org/docs/porting/connecting_cpp_and_javascript/embind.html) allows developers to write C++
code and directly interact with the code from JS.

This is done by registering enums, functions, classes, arrays, objects, vectors and maps from C++. When the compiled
WebAssembly initializes, it will register all those types in the host using host function calls.
Due to these registrations, the host knows how to encode/decode values to communicate with the guest.

## Wazero implementation

This implementation is trying to be a 1-on-1 implementation of the JS version in Emscripten itself, so that the same
codebase can be used for both the web and WASI WebAssembly.
The difference between this implementation and the Emscripten implementation is that this implementation tries to be as
strict as possible regarding the types that are encoded/decoded, while in the Emscripten implementation a lot is trusted
to the JS VM to cast between types, something we can't do in Go.

## Compiling with Emscripten to WebAssembly with Embind

Be sure to read the [documentation](https://emscripten.org/docs/porting/connecting_cpp_and_javascript/embind.html) to
get to know how Embind works. The most basic version to compile something with Embind is:

```shell
emcc -sERROR_ON_UNDEFINED_SYMBOLS=0 -sEXPORTED_FUNCTIONS="_free,_malloc" -g embind_example.cpp -o embind.wasm -lembind --no-entry
```

It is very important to include `-lembind` to the command and export the functions `_free` and `__malloc`, if these
functions are not available, this package won't work.

## Attaching the Embind Engine to the context

The Embind Engine allows itself to be attached to a context value so that it can be attached to the Wazero runtime.
This is neccesary so that the guest side can register itself with the Engine to notify it of all the available Embind
components.

Here is an example to setup a basic Wazero runtime with Embind:
<details>
  <summary>main.go</summary>

```go
package main

import (
    "context"
    "log"

    "github.com/jerbob92/wazero-emscripten-embind"
    "github.com/tetratelabs/wazero"
    "github.com/tetratelabs/wazero/imports/emscripten"
    "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed wasm/embind.wasm
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
        WithName("")

    ctx = engine.Attach(ctx)
    _, err = r.InstantiateModule(ctx, compiledModule, moduleConfig)
    if err != nil {
        log.Fatal(err)
    }

    // If you have a generated package, you have to attach it to the engine to
    // register the generated values/types with the Engine.
    err = generated.Attach(engine)
    if err != nil {
        log.Fatal(err)
    }
}
```
</details>

You can find more examples in the examples directory.

## Code generator

This project includes a code generator that will automatically generate typed code based on a given WASM file that has
Embind
instructions in it. You can generate the code like this:

First create a file in a new package, let's call it `generated/generated.go` for now. Add the following to the file:

```go
//go:generate go run github.com/jerbob92/wazero-emscripten-embind/generator -wasm=../wasm/embind.wasm
package generated
```

Where `../wasm/embind.wasm` is the path to the WASM file, relative to the file `generated.go`.

You can now run the command from the `generated` folder to make it generate the typed Go code:

```shell
go generate
```

Or from the project root:

```shell
go generate ./...
```

## Using Embind from Go

The easiest way to call Embind from Go would be to use the generator, but it's also possible to do things directly using
the Engine, the generated code is basically a wrapper around the Engine.

Here are a few examples:

```go
// Calling an exposed symbol (function) called returnRawData with a string argument.
imageData, err := engine.CallPublicSymbol(ctx, "returnRawData", "image")

// Creating a new instance of the class MyClass using the values 5 and 10 in the constructor.
newClass, err := engine.CallPublicSymbol(ctx, "MyClass", 50, 10)
```

## Using Go from Embind

This package allows you to use Go code directly from Embind
using [Emval](https://emscripten.org/docs/api_reference/val.h.html), in the same way that would work in JS, the only
difference is that in JS, the full global namespace is available, while in Go you specifically have to expose things to
Embind to be able to access them.

You can do the following things with this:

* Call methods on structs
* Set/Get properties on structs
* Create new instances of structs
* Share arbitrary data like strings and integers

For example, given the following Go code:

```go
package main

import (
	"github.com/jerbob92/wazero-emscripten-embind"
	"github.com/tetratelabs/wazero"
	"log"
)

type testStruct struct {
	Property1 string `embind_arg:"0" embind_property:"propone"`
	Property2 string `embind_property:"proptwo"`
	Property3 string
}

func (ts *testStruct) Trigger() {
	log.Printf("Triggering %s %s %s on testStruct", ts.Property1, ts.Property2, ts.Property3)
}

func main() {
	// Initialize Wazero runtime and Embind engine ...
	engine.RegisterEmvalSymbol("testStruct", &testStruct{})
}
```

You can then do the following on the C++ side:

```cpp
val testStruct = val::global("testStruct");
val newStruct = testStruct.new_("valueInProperty1");
newStruct.set("proptwo", val("valueInProperty2"));
newStruct.set("Property3", val("valueInProperty3"));
newStruct.call<void>("Trigger");
```

A few things to note:

* You can return structs and then also set/get properties and call methods on that
* If your function is void in C++, you can return nothing or an error in Go
* If your function has a return in C++, you can return something, or something and an error, where the error has to be
  the second return
* If your function returns an error, the whole call where the Emval call originated from will fail
* You can use the `embind_arg` tag to tell the Engine which argument index should end up in which property in
  case `.new_()` is used on the C++ side
* You can implement the `embind.EmvalConstructor` interface on the struct to make your own constructor
* You can use the `embind_property` tag to tell the Engine which property should be access when a set or get is done in
  C++
* You can implement the `embind.EmvalFunctionMapper` interface on the struct to map function calls on your struct based
  on the arguments (and/or length) and name
