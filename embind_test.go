package embind_test

import (
	"context"
	"log"
	"os"
	"testing"

	embind_external "github.com/jerbob92/wazero-emscripten-embind"
	"github.com/jerbob92/wazero-emscripten-embind/generator/generator"
	embind "github.com/jerbob92/wazero-emscripten-embind/internal"

	"github.com/jerbob92/wazero-emscripten-embind/types"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/emscripten"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/tetratelabs/wazero/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestEmbind(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Embind Suite")
}

var ctx = context.Background()
var engine embind_external.Engine
var runtime wazero.Runtime
var mod api.Module
var wasmData []byte

var _ = BeforeSuite(func() {
	wasm, err := os.ReadFile("./testdata/wasm/tests.wasm")
	if err != nil {
		Expect(err).To(BeNil())
		return
	}

	wasmData = wasm

	runtimeConfig := wazero.NewRuntimeConfig()
	runtime = wazero.NewRuntimeWithConfig(ctx, runtimeConfig)

	if _, err := wasi_snapshot_preview1.Instantiate(ctx, runtime); err != nil {
		Expect(err).To(BeNil())
		return
	}

	compiledModule, err := runtime.CompileModule(ctx, wasm)
	if err != nil {
		log.Fatal(err)
	}

	builder := runtime.NewHostModuleBuilder("env")

	emscriptenExporter, err := emscripten.NewFunctionExporterForModule(compiledModule)
	if err != nil {
		Expect(err).To(BeNil())
		return
	}

	emscriptenExporter.ExportFunctions(builder)

	engine = embind_external.CreateEngine(embind_external.NewConfig())

	embindExporter := engine.NewFunctionExporterForModule(compiledModule)
	err = embindExporter.ExportFunctions(builder)
	if err != nil {
		Expect(err).To(BeNil())
		return
	}

	_, err = builder.Instantiate(ctx)
	if err != nil {
		Expect(err).To(BeNil())
		return
	}

	moduleConfig := wazero.NewModuleConfig().
		WithStartFunctions("_initialize").
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithName("")

	ctx = engine.Attach(ctx)
	mod, err = runtime.InstantiateModule(ctx, compiledModule, moduleConfig)
	if err != nil {
		Expect(err).To(BeNil())
		return
	}
})

var _ = AfterSuite(func() {
	runtime.Close(ctx)
})

var _ = Describe("Calling embind functions", Label("library"), func() {
	When("the function is being called", func() {
		It("gives an error on an invalid argument count", func() {
			res, err := engine.CallPublicSymbol(ctx, "bool_return_bool", 1, 2)
			Expect(err).To(Not(BeNil()))
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("function bool_return_bool called with 2 argument(s), expected 1 arg(s)"))
			}
			Expect(res).To(BeNil())
		})
		Context("the return type is bool", func() {
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "bool_return_true")
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_false")
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", false)
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", true)
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", int(0))
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", int(1))
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", uint(0))
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", uint(1))
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", int8(0))
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", int8(1))
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", uint8(0))
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", uint8(1))
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", int16(0))
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", int16(1))
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", uint16(0))
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", uint16(1))
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", int32(0))
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", int32(1))
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", uint32(0))
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", uint32(1))
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", uint16(1))
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", int64(0))
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", int64(1))
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", uint64(0))
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", uint64(1))
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", float32(0))
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", float32(1))
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", float64(0))
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", float64(1))
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", "")
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", "123")
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_bool", struct {
				}{})
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())
			})
		})

		Context("the return type is void", func() {
			It("has the correct return type", func() {
				res, err := engine.CallPublicSymbol(ctx, "float_return_void", float32(123))
				Expect(err).To(BeNil())
				Expect(res).To(BeNil())
			})
		})

		Context("the return type is float", func() {
			It("gives an error on an invalid input", func() {
				res, err := engine.CallPublicSymbol(ctx, "float_return_float", 1)
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (float): value must be of type float32"))
				}
				Expect(res).To(BeNil())
			})
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "float_return_float", float32(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(float32(9)))
			})
		})

		Context("the return type is double", func() {
			It("gives an error on an invalid input", func() {
				res, err := engine.CallPublicSymbol(ctx, "double_return_double", 1)
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (double): value must be of type float64"))
				}
				Expect(res).To(BeNil())
			})
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "double_return_double", float64(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(float64(9)))
			})
		})

		Context("the return type is int", func() {
			It("gives an error on an invalid input", func() {
				res, err := engine.CallPublicSymbol(ctx, "int_return_int", 1)
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (int): value must be of type int32"))
				}
				Expect(res).To(BeNil())
			})
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "int_return_int", int32(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(int32(9)))
			})
		})

		Context("the return type is char", func() {
			It("gives an error on an invalid input", func() {
				res, err := engine.CallPublicSymbol(ctx, "char_return_char", 1)
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (char): value must be of type int8, is int"))
				}
				Expect(res).To(BeNil())
			})
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "char_return_char", int8(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(int8(3)))
			})
		})

		Context("the return type is long", func() {
			It("gives an error on an invalid input", func() {
				res, err := engine.CallPublicSymbol(ctx, "long_return_long", 1)
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (long): value must be of type int32, is int"))
				}
				Expect(res).To(BeNil())
			})
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "long_return_long", int32(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(int32(6)))
			})
		})

		Context("the return type is short", func() {
			It("gives an error on an invalid input", func() {
				res, err := engine.CallPublicSymbol(ctx, "short_return_short", 1)
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (short): value must be of type int16, is int"))
				}
				Expect(res).To(BeNil())
			})
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "short_return_short", int16(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(int16(6)))
			})
		})

		Context("the return type is unsigned char", func() {
			It("gives an error on an invalid input", func() {
				res, err := engine.CallPublicSymbol(ctx, "uchar_return_uchar", 1)
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (unsigned char): value must be of type uint8, is int"))
				}
				Expect(res).To(BeNil())
			})
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "uchar_return_uchar", uint8(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(uint8(3)))
			})
		})

		Context("the return type is unsigned int", func() {
			It("gives an error on an invalid input", func() {
				res, err := engine.CallPublicSymbol(ctx, "uint_return_uint", 1)
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (unsigned int): value must be of type uint32, is int"))
				}
				Expect(res).To(BeNil())
			})
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "uint_return_uint", uint32(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(uint32(6)))
			})
		})

		Context("the return type is unsigned long", func() {
			It("gives an error on an invalid input", func() {
				res, err := engine.CallPublicSymbol(ctx, "ulong_return_ulong", 1)
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (unsigned long): value must be of type uint32, is int"))
				}
				Expect(res).To(BeNil())
			})
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "ulong_return_ulong", uint32(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(uint32(6)))
			})
		})

		Context("the return type is unsigned short", func() {
			It("gives an error on an invalid input", func() {
				res, err := engine.CallPublicSymbol(ctx, "ushort_return_ushort", 1)
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (unsigned short): value must be of type uint16, is int"))
				}
				Expect(res).To(BeNil())
			})
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "ushort_return_ushort", uint16(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(uint16(6)))
			})
		})

		Context("the return type is long long", func() {
			It("gives an error on an invalid input", func() {
				res, err := engine.CallPublicSymbol(ctx, "longlong_return_longlong", 1)
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (int64_t): value must be of type int64"))
				}
				Expect(res).To(BeNil())
			})
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "longlong_return_longlong", int64(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(int64(6)))
			})
		})

		Context("the return type is unsigned long long", func() {
			It("gives an error on an invalid input", func() {
				res, err := engine.CallPublicSymbol(ctx, "ulonglong_return_ulonglong", 1)
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (uint64_t): value must be of type uint64"))
				}
				Expect(res).To(BeNil())
			})
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "ulonglong_return_ulonglong", uint64(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(uint64(6)))
			})
		})

		Context("the return type is std string", func() {
			It("gives an error on an invalid input", func() {
				res, err := engine.CallPublicSymbol(ctx, "std_string_return_std_string", 1)
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (std::string): value must be of type string"))
				}
				Expect(res).To(BeNil())
			})
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "std_string_return_std_string", "embind")
				Expect(err).To(BeNil())
				Expect(res).To(Equal("Hello there embind"))
			})
		})

		Context("the return type is std wstring", func() {
			It("gives an error on an invalid input", func() {
				res, err := engine.CallPublicSymbol(ctx, "std_wstring_return_std_wstring", 1)
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (std::wstring): input must be a string, was int"))
				}
				Expect(res).To(BeNil())
			})
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "std_wstring_return_std_wstring", "embind")
				Expect(err).To(BeNil())
				Expect(res).To(Equal("Hello there embind"))
			})
		})

		Context("the return type is std u16string", func() {
			It("gives an error on an invalid input", func() {
				res, err := engine.CallPublicSymbol(ctx, "std_u16string_return_std_u16string", 1)
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (std::u16string): input must be a string, was int"))
				}
				Expect(res).To(BeNil())
			})
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "std_u16string_return_std_u16string", "embind")
				Expect(err).To(BeNil())
				Expect(res).To(Equal("Hello there embind"))
			})
		})

		Context("the return type is a vector", func() {
			It("has the correct return value", func() {
				res, err := engine.CallPublicSymbol(ctx, "return_vector")
				Expect(err).To(BeNil())
				Expect(res).To(Not(BeNil()))
				Expect(res).To(BeAssignableToTypeOf(&embind.ClassBase{}))
				if obj, ok := res.(*embind.ClassBase); ok {
					size, err := obj.CallInstanceMethod(ctx, obj, "size")
					Expect(err).To(BeNil())
					Expect(size).To(Equal(uint32(10)))

					_, err = obj.CallInstanceMethod(ctx, obj, "resize", uint32(12), int32(1))
					Expect(err).To(BeNil())

					size, err = obj.CallInstanceMethod(ctx, obj, "size")
					Expect(err).To(BeNil())
					Expect(size).To(Equal(uint32(12)))

					val, err := obj.CallInstanceMethod(ctx, obj, "get", uint32(1))
					Expect(err).To(BeNil())
					Expect(val).To(Equal(int32(1)))

					_, err = obj.CallInstanceMethod(ctx, obj, "set", uint32(1), int32(2))
					Expect(err).To(BeNil())

					val, err = obj.CallInstanceMethod(ctx, obj, "get", uint32(1))
					Expect(err).To(BeNil())
					Expect(val).To(Equal(int32(2)))
				}
			})
		})

		Context("the return type is a map", func() {
			It("has the correct return value", func() {
				res, err := engine.CallPublicSymbol(ctx, "return_map")
				Expect(err).To(BeNil())
				Expect(res).To(Not(BeNil()))
				Expect(res).To(BeAssignableToTypeOf(&embind.ClassBase{}))
				if obj, ok := res.(*embind.ClassBase); ok {
					size, err := obj.CallInstanceMethod(ctx, obj, "size")
					Expect(err).To(BeNil())
					Expect(size).To(Equal(uint32(1)))

					val, err := obj.CallInstanceMethod(ctx, obj, "get", int32(10))
					Expect(err).To(BeNil())
					Expect(val).To(Equal("This is a string."))

					val, err = obj.CallInstanceMethod(ctx, obj, "get", int32(1))
					Expect(err).To(BeNil())
					Expect(val).To(Equal(types.Undefined))
				}
			})
		})

		Context("the return type is a memory view", func() {
			Context("of type char", func() {
				It("has the correct values", func() {
					res, err := engine.CallPublicSymbol(ctx, "get_memory_view_char")
					Expect(err).To(BeNil())
					Expect(res).To(Equal([]int8{0, 1, 2, 3, 4, 5}))
				})
			})

			Context("of type uchar", func() {
				It("has the correct values", func() {
					res, err := engine.CallPublicSymbol(ctx, "get_memory_view_unsigned_char")
					Expect(err).To(BeNil())
					Expect(res).To(Equal([]uint8{0, 1, 2, 3, 4, 5}))
				})
			})

			Context("of type int", func() {
				It("has the correct values", func() {
					res, err := engine.CallPublicSymbol(ctx, "get_memory_view_int")
					Expect(err).To(BeNil())
					Expect(res).To(Equal([]int32{0, 1, 2, 3, 4, 5}))
				})
			})

			Context("of type uint", func() {
				It("has the correct values", func() {
					res, err := engine.CallPublicSymbol(ctx, "get_memory_view_unsigned_int")
					Expect(err).To(BeNil())
					Expect(res).To(Equal([]uint32{0, 1, 2, 3, 4, 5}))
				})
			})

			Context("of type long", func() {
				It("has the correct values", func() {
					res, err := engine.CallPublicSymbol(ctx, "get_memory_view_long")
					Expect(err).To(BeNil())
					Expect(res).To(Equal([]int32{0, 1, 2, 3, 4, 5}))
				})
			})

			Context("of type unsigned long", func() {
				It("has the correct values", func() {
					res, err := engine.CallPublicSymbol(ctx, "get_memory_view_unsigned_long")
					Expect(err).To(BeNil())
					Expect(res).To(Equal([]uint32{0, 1, 2, 3, 4, 5}))
				})
			})

			Context("of type short", func() {
				It("has the correct values", func() {
					res, err := engine.CallPublicSymbol(ctx, "get_memory_view_short")
					Expect(err).To(BeNil())
					Expect(res).To(Equal([]int16{0, 1, 2, 3, 4, 5}))
				})
			})

			Context("of type ushort", func() {
				It("has the correct values", func() {
					res, err := engine.CallPublicSymbol(ctx, "get_memory_view_unsigned_short")
					Expect(err).To(BeNil())
					Expect(res).To(Equal([]uint16{0, 1, 2, 3, 4, 5}))
				})
			})

			Context("of type longlong", func() {
				It("has the correct values", func() {
					res, err := engine.CallPublicSymbol(ctx, "get_memory_view_longlong")
					Expect(err).To(BeNil())
					Expect(res).To(Equal([]int64{0, 1, 2, 3, 4, 5}))
				})
			})

			Context("of type unsigned longlong", func() {
				It("has the correct values", func() {
					res, err := engine.CallPublicSymbol(ctx, "get_memory_view_unsigned_longlong")
					Expect(err).To(BeNil())
					Expect(res).To(Equal([]uint64{0, 1, 2, 3, 4, 5}))
				})
			})

			Context("of type double", func() {
				It("has the correct values", func() {
					res, err := engine.CallPublicSymbol(ctx, "get_memory_view_double")
					Expect(err).To(BeNil())
					Expect(res).To(Equal([]float64{0, 1, 2, 3, 4, 5}))
				})
			})

			Context("of type float", func() {
				It("has the correct values", func() {
					res, err := engine.CallPublicSymbol(ctx, "get_memory_view_float")
					Expect(err).To(BeNil())
					Expect(res).To(Equal([]float32{0, 1, 2, 3, 4, 5}))
				})
			})
		})
		Context("when an overload table is used", func() {
			It("gives an error on an unknown number of arguments", func() {
				res, err := engine.CallPublicSymbol(ctx, "function_overload", 1, 2, 3)
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("function 'function_overload' called with an invalid number of arguments (3) - expects one of (0, 1)"))
				}
				Expect(res).To(BeNil())
			})

			It("gives an error on a wrong argument type", func() {
				res, err := engine.CallPublicSymbol(ctx, "function_overload", "test")
				Expect(err).To(Not(BeNil()))
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("function_overload: could not get wire type of argument 0 (int): value must be of type int32, is string"))
				}
				Expect(res).To(BeNil())
			})

			It("picks the correct item from the overload table", func() {
				res, err := engine.CallPublicSymbol(ctx, "function_overload")
				Expect(err).To(BeNil())
				Expect(res).To(Equal(int32(1)))

				res, err = engine.CallPublicSymbol(ctx, "function_overload", int32(12))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(int32(2)))
			})
		})
	})
})

var _ = Describe("Using embind constants", Label("library"), func() {
	When("the constants are being registered", func() {
		It("has the correct values", func() {
			var constantsMap = map[string]any{}
			constants := engine.GetConstants()
			for i := range constants {
				constantsMap[constants[i].Name()] = constants[i].Value()
			}

			Expect(constantsMap).To(HaveKeyWithValue("SOME_CONSTANT_1", false))
			Expect(constantsMap).To(HaveKeyWithValue("SOME_CONSTANT_2", float32(2)))
			Expect(constantsMap).To(HaveKeyWithValue("SOME_CONSTANT_3", float64(3)))
			Expect(constantsMap).To(HaveKeyWithValue("SOME_CONSTANT_4", int32(4)))
			Expect(constantsMap).To(HaveKeyWithValue("SOME_CONSTANT_5", "TestString"))
			Expect(constantsMap).To(HaveKeyWithValue("SOME_CONSTANT_6", int8(67)))
			Expect(constantsMap).To(HaveKeyWithValue("SOME_CONSTANT_7", int64(7)))
			Expect(constantsMap).To(HaveKeyWithValue("SOME_CONSTANT_8", uint16(8)))
			Expect(constantsMap).To(HaveKeyWithValue("SOME_CONSTANT_9", uint32(9)))
			Expect(constantsMap).To(HaveKeyWithValue("SOME_CONSTANT_10", uint16(10)))
			Expect(constantsMap).To(HaveKeyWithValue("SOME_CONSTANT_11", uint8(11)))
			Expect(constantsMap).To(HaveKeyWithValue("SOME_CONSTANT_12", uint32(12)))
			Expect(constantsMap).To(HaveKeyWithValue("SOME_CONSTANT_13", "TestWideString"))
			Expect(constantsMap).To(HaveKeyWithValue("SOME_CONSTANT_14", true))
			Expect(constantsMap).To(HaveKeyWithValue("SOME_CONSTANT_15", uint64(15)))
			Expect(constantsMap).To(HaveKeyWithValue("hasUnboundTypeNames", true))
			Expect(constantsMap).To(HaveLen(16))
		})
	})
})

var _ = Describe("Using embind enums", Label("library"), func() {
	When("the enums are being registered", func() {
		It("has the correct values", func() {
			var enumsMap = map[string]map[string]any{}
			enums := engine.GetEnums()
			for i := range enums {
				values := map[string]any{}

				enumValues := enums[i].Values()
				for valI := range enumValues {
					values[enumValues[valI].Name()] = enumValues[valI].Value()
				}

				enumsMap[enums[i].Name()] = values
			}

			Expect(enumsMap).To(HaveKeyWithValue("OldStyle", map[string]any{
				"ONE": uint32(0),
				"TWO": uint32(1),
			}))
			Expect(enumsMap["OldStyle"]).To(HaveLen(2))
			Expect(enumsMap).To(HaveKeyWithValue("NewStyle", map[string]any{
				"ONE": int32(0),
				"TWO": int32(1),
			}))
			Expect(enumsMap["NewStyle"]).To(HaveLen(2))
		})

		It("can be encoded and decoded", func() {
			res, err := engine.CallPublicSymbol(ctx, "enum_in_enum_out", int32(0))
			Expect(err).To(BeNil())
			Expect(res).To(Equal(uint32(1)))
		})
	})
})

var _ = Describe("Using embind structs", Label("library"), func() {
	When("using the structs", func() {
		It("can be decoded with an array as input", func() {
			res, err := engine.CallPublicSymbol(ctx, "findPersonAtLocation", []any{float32(1), float32(2)})
			Expect(err).To(BeNil())
			Expect(res).To(Equal(map[string]any{
				"name": "123",
				"age":  int32(12),
				"structArray": map[string]any{
					"field": []any{int32(1), int32(2)},
				},
			}))
		})

		It("can be encoded with an array and struct as input", func() {
			res, err := engine.CallPublicSymbol(ctx, "setPersonAtLocation", []any{float32(1), float32(2)}, map[string]any{
				"name": "123",
				"age":  int32(12),
				"structArray": map[string]any{
					"field": []any{int32(1), int32(2)},
				},
			})
			Expect(err).To(BeNil())
			Expect(res).To(BeNil())
		})

		It("gives an error when not giving all fields as input", func() {
			res, err := engine.CallPublicSymbol(ctx, "setPersonAtLocation", []any{float32(1), float32(2)}, map[string]any{
				"fakefield": "123",
			})
			Expect(err).To(Not(BeNil()))
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("missing field: name"))
			}
			Expect(res).To(BeNil())
		})

		It("gives an error when an incomplete number of array elements is given", func() {
			res, err := engine.CallPublicSymbol(ctx, "setPersonAtLocation", []any{float32(1), float32(2)}, map[string]any{
				"name": "123",
				"age":  int32(12),
				"structArray": map[string]any{
					"field": []any{int32(1)},
				},
			})
			Expect(err).To(Not(BeNil()))
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("incorrect number of tuple elements for array_int_2: expected=2, actual=1"))
			}
			Expect(res).To(BeNil())
		})
	})
})

type webkitAudioContextOscillatorFrequency struct {
	Value uint64 `embind_property:"value"`
}

type webkitAudioContextOscillator struct {
	Typing     string                                 `embind_property:"type"`
	Frequencdy *webkitAudioContextOscillatorFrequency `embind_property:"frequency"`
}

func (waco *webkitAudioContextOscillator) Connect(destination string) error {
	return nil
}

func (waco *webkitAudioContextOscillator) Start(destination int32) error {
	return nil
}

func (waco *webkitAudioContextOscillator) MapFunction(name string, returnType string, argTypes []string) (string, error) {
	if name == "start" {
		return "Start", nil
	}
	return "", nil
}

type webkitAudioContext struct {
	Destination string `embind_arg:"0"`
}

func (was *webkitAudioContext) CreateOscillator() (*webkitAudioContextOscillator, error) {
	return &webkitAudioContextOscillator{
		Frequencdy: &webkitAudioContextOscillatorFrequency{},
	}, nil
}

type ICreateOscillator interface {
	CreateOscillator() *webkitAudioContextOscillator
}

var _ = Describe("Using embind emval", Label("library"), func() {
	When("using the Go struct mapping", func() {
		It("fails when no struct is mapped", func() {
			_, err := engine.CallPublicSymbol(ctx, "doEmval")
			Expect(err).To(Not(BeNil()))
		})

		It("can map the struct", func() {
			c2 := &webkitAudioContext{}
			err := engine.RegisterEmvalSymbol("webkitAudioContext", c2)
			Expect(err).To(BeNil())
		})

		It("gives an error when a non pointer struct is already mapped", func() {
			c2 := &webkitAudioContext{}
			err := engine.RegisterEmvalSymbol("webkitAudioContext", c2)
			Expect(err).To(Not(BeNil()))
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("could not register symbol webkitAudioContext, already registered as type *embind_test.webkitAudioContext"))
			}
		})

		It("can use the full struct from C++", func() {
			res, err := engine.CallPublicSymbol(ctx, "doEmval")
			Expect(err).To(BeNil())
			Expect(res).To(Equal(`No global AudioContext, trying webkitAudioContext
Got an AudioContext
Configuring oscillator
Playing
All done!
`))
		})
	})
})

var _ = Describe("Using embind classes", Label("library"), func() {
	When("Constructing a new class", func() {
		It("fails when an invalid number of arguments is given", func() {
			res, err := engine.CallPublicSymbol(ctx, "MyClass")
			Expect(err).To(Not(BeNil()))
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("with invalid number of parameters (0) - expected (1 or 2) parameters instead"))
			}
			Expect(res).To(BeNil())
		})
		It("fails when an invalid argument is given", func() {
			res, err := engine.CallPublicSymbol(ctx, "MyClass", float64(123))
			Expect(err).To(Not(BeNil()))
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (int): value must be of type int32, is float64"))
			}
			Expect(res).To(BeNil())
		})

		It("succeeds to construct with both overloads", func() {
			res, err := engine.CallPublicSymbol(ctx, "MyClass", int32(123))
			Expect(err).To(BeNil())
			Expect(res).To(Not(BeNil()))
			Expect(res).To(BeAssignableToTypeOf(&embind.ClassBase{}))

			res, err = engine.CallPublicSymbol(ctx, "MyClass", int32(123), "test123")
			Expect(err).To(BeNil())
			Expect(res).To(Not(BeNil()))
			Expect(res).To(BeAssignableToTypeOf(&embind.ClassBase{}))
		})

		Context("when the class has been constructed", func() {
			var myClass *embind.ClassBase
			BeforeEach(func() {
				res, err := engine.CallPublicSymbol(ctx, "MyClass", int32(123), "test")
				Expect(err).To(BeNil())
				Expect(res).To(Not(BeNil()))
				Expect(res).To(BeAssignableToTypeOf(&embind.ClassBase{}))
				if obj, ok := res.(*embind.ClassBase); ok {
					myClass = obj
				}
			})

			AfterEach(func() {
				if myClass != nil {
					err := myClass.DeleteInstance(ctx, myClass)
					Expect(err).To(BeNil())
				}
			})

			Context("when calling functions", func() {
				It("gives an error on an unknown function", func() {
					res, err := myClass.CallInstanceMethod(ctx, myClass, "unknown", 1, 2, 3)
					Expect(err).To(Not(BeNil()))
					if err != nil {
						Expect(err.Error()).To(ContainSubstring("method unknown is not found"))
					}
					Expect(res).To(BeNil())
				})

				It("gives an error on a function with a wrong argument count", func() {
					res, err := myClass.CallInstanceMethod(ctx, myClass, "combineY", 1, 2, 3)
					Expect(err).To(Not(BeNil()))
					if err != nil {
						Expect(err.Error()).To(ContainSubstring("called with 3 argument(s), expected 1 arg(s)"))
					}
					Expect(res).To(BeNil())
				})

				It("gives an error on a function with a wrong argument", func() {
					res, err := myClass.CallInstanceMethod(ctx, myClass, "combineY", 1)
					Expect(err).To(Not(BeNil()))
					if err != nil {
						Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (std::string): value must be of type string"))
					}
					Expect(res).To(BeNil())
				})

				It("can call the method correctly", func() {
					res, err := myClass.CallInstanceMethod(ctx, myClass, "combineY", "hello ")
					Expect(err).To(BeNil())
					Expect(res).To(Equal("hello test"))
				})

				Context("that have overloads", func() {
					It("fails when giving an invalid overload", func() {
						res, err := myClass.CallInstanceMethod(ctx, myClass, "incrementX", 1, 2, 3)
						Expect(err).To(Not(BeNil()))
						if err != nil {
							Expect(err.Error()).To(ContainSubstring("called with an invalid number of arguments (3) - expects one of (0, 1)"))
						}
						Expect(res).To(BeNil())
					})

					It("works with each of the overloads", func() {
						res, err := myClass.CallInstanceMethod(ctx, myClass, "incrementX", int32(1))
						Expect(err).To(BeNil())
						Expect(res).To(BeNil())

						res, err = myClass.CallInstanceMethod(ctx, myClass, "incrementX")
						Expect(err).To(BeNil())
						Expect(res).To(BeNil())
					})
				})
			})
			Context("when calling setters/getters", func() {
				It("gives an error on an unknown property", func() {
					res, err := myClass.GetInstanceProperty(ctx, myClass, "test")
					Expect(err).To(Not(BeNil()))
					if err != nil {
						Expect(err.Error()).To(ContainSubstring("property test is not found"))
					}
					Expect(res).To(BeNil())

					err = myClass.SetInstanceProperty(ctx, myClass, "test", 123)
					Expect(err).To(Not(BeNil()))
					if err != nil {
						Expect(err.Error()).To(ContainSubstring("property test is not found"))
					}
					Expect(res).To(BeNil())
				})

				It("gives an error when setting on a readonly property", func() {
					err := myClass.SetInstanceProperty(ctx, myClass, "y", "")
					Expect(err).To(Not(BeNil()))
					if err != nil {
						Expect(err.Error()).To(ContainSubstring("is read-only"))
					}
				})

				It("gives an error when setting with a wrong argument", func() {
					err := myClass.SetInstanceProperty(ctx, myClass, "x", "")
					Expect(err).To(Not(BeNil()))
					if err != nil {
						Expect(err.Error()).To(ContainSubstring("value must be of type int32, is string"))
					}
				})

				It("allows setting and getting a property", func() {
					err := myClass.SetInstanceProperty(ctx, myClass, "x", int32(3))
					Expect(err).To(BeNil())

					res, err := myClass.GetInstanceProperty(ctx, myClass, "x")
					Expect(err).To(BeNil())
					Expect(res).To(Equal(int32(3)))
				})

				It("allows getting a property", func() {
					res, err := myClass.GetInstanceProperty(ctx, myClass, "x")
					Expect(err).To(BeNil())
					Expect(res).To(Equal(int32(123)))
				})
			})

			Context("when calling static methods", func() {
				It("gives an error on an unknown class", func() {
					res, err := engine.CallStaticClassMethod(ctx, "MyClass123", "test")
					Expect(err).To(Not(BeNil()))
					if err != nil {
						Expect(err.Error()).To(ContainSubstring("could not find class MyClass123"))
					}
					Expect(res).To(BeNil())
				})

				It("gives an error on an unknown method", func() {
					res, err := engine.CallStaticClassMethod(ctx, "MyClass", "test")
					Expect(err).To(Not(BeNil()))
					if err != nil {
						Expect(err.Error()).To(ContainSubstring("could not find method test on class MyClass"))
					}
					Expect(res).To(BeNil())
				})

				It("gives an error on an invalid argument", func() {
					res, err := engine.CallStaticClassMethod(ctx, "MyClass", "getStringFromInstance", 123)
					Expect(err).To(Not(BeNil()))
					if err != nil {
						Expect(err.Error()).To(ContainSubstring("could not get wire type of argument 0 (MyClass): invalid MyClass, check whether you constructed it properly through embind, the given value is a int"))
					}
					Expect(res).To(BeNil())
				})

				It("gives an error on an invalid argument count", func() {
					res, err := engine.CallStaticClassMethod(ctx, "MyClass", "getStringFromInstance", 1, 2)
					Expect(err).To(Not(BeNil()))
					if err != nil {
						Expect(err.Error()).To(ContainSubstring("function MyClass.getStringFromInstance called with 2 argument(s), expected 1 arg(s)"))
					}
					Expect(res).To(BeNil())
				})

				It("allows calling a static method", func() {
					res, err := engine.CallStaticClassMethod(ctx, "MyClass", "getStringFromInstance", myClass)
					Expect(err).To(BeNil())
					Expect(res).To(Equal("test"))
				})
			})
		})
	})
	When("A class is mapped to a Go struct", func() {
		It("fails to map when the class doesn't embed the class base", func() {
			type ClassMyClass struct{}
			err := engine.RegisterClass("MyClass", &ClassMyClass{})
			Expect(err).To(Not(BeNil()))
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("it does not embed embind.ClassBase"))
			}
		})
		It("fails to map when the class isn't passed as pointer", func() {
			type ClassMyClass struct {
				embind_external.ClassBase
			}
			err := engine.RegisterClass("MyClass", ClassMyClass{})
			Expect(err).To(Not(BeNil()))
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("given value should be a pointer type"))
			}
		})
		It("succeeds to construct with both overloads", func() {
			type ClassMyClass struct {
				embind_external.ClassBase
			}
			err := engine.RegisterClass("MyClass", &ClassMyClass{})
			Expect(err).To(BeNil())

			res, err := engine.CallPublicSymbol(ctx, "MyClass", int32(123))
			Expect(err).To(BeNil())
			Expect(res).To(Not(BeNil()))
			Expect(res).To(BeAssignableToTypeOf(&ClassMyClass{}))

			res, err = engine.CallPublicSymbol(ctx, "MyClass", int32(123), "test123")
			Expect(err).To(BeNil())
			Expect(res).To(Not(BeNil()))
			Expect(res).To(BeAssignableToTypeOf(&ClassMyClass{}))

			passThrough, err := engine.CallPublicSymbol(ctx, "passThrough", res)
			Expect(err).To(BeNil())
			Expect(passThrough).To(Not(BeNil()))
			Expect(passThrough).To(BeAssignableToTypeOf(&ClassMyClass{}))
			Expect(passThrough).To(Equal(res))
		})

		It("errors when it is already mapped", func() {
			type ClassMyClass struct {
				embind_external.ClassBase
			}
			err := engine.RegisterClass("MyClass", &ClassMyClass{})
			Expect(err).To(Not(BeNil()))
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("could not register class MyClass, already registered as type *embind_test.ClassMyClass"))
			}
		})
	})
})

var _ = Describe("Using the generator", Label("generator"), func() {
	When("generating the code", func() {
		It("succeeds generating the code", func() {
			err := generator.Generate("./tests/generated", "./tests/generated/generate.go", wasmData, "_initialize")
			Expect(err).To(BeNil())
		})
	})
})
