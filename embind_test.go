package embind

import (
	"context"
	"log"
	"os"
	"testing"

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
var engine Engine
var runtime wazero.Runtime
var mod api.Module

var _ = BeforeSuite(func() {
	wasm, err := os.ReadFile("./testdata/wasm/tests.wasm")
	if err != nil {
		Expect(err).To(BeNil())
		return
	}

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

	engine = CreateEngine(NewConfig())

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
		Context("the return type is bool", func() {
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "bool_return_true")
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				res, err = engine.CallPublicSymbol(ctx, "bool_return_false")
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())
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
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "float_return_float", float32(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(float32(9)))
			})
		})

		Context("the return type is double", func() {
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "double_return_double", float64(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(float64(9)))
			})
		})

		Context("the return type is int", func() {
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "int_return_int", int32(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(int32(9)))
			})
		})

		Context("the return type is char", func() {
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "char_return_char", int8(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(int8(3)))
			})
		})

		Context("the return type is long", func() {
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "long_return_long", int32(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(int32(6)))
			})
		})

		Context("the return type is short", func() {
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "short_return_short", int16(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(int16(6)))
			})
		})

		Context("the return type is unsigned char", func() {
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "uchar_return_uchar", uint8(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(uint8(3)))
			})
		})

		Context("the return type is unsigned int", func() {
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "uint_return_uint", uint32(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(uint32(6)))
			})
		})

		Context("the return type is unsigned long", func() {
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "ulong_return_ulong", uint32(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(uint32(6)))
			})
		})

		Context("the return type is unsigned short", func() {
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "ushort_return_ushort", uint16(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(uint16(6)))
			})
		})

		Context("the return type is long long", func() {
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "longlong_return_longlong", int64(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(int64(6)))
			})
		})

		Context("the return type is unsigned long long", func() {
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "ulonglong_return_ulonglong", uint64(3))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(uint64(6)))
			})
		})

		Context("the return type is std string", func() {
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "std_string_return_std_string", "embind")
				Expect(err).To(BeNil())
				Expect(res).To(Equal("Hello there embind"))
			})
		})

		Context("the return type is std wstring", func() {
			It("has the correct return values", func() {
				res, err := engine.CallPublicSymbol(ctx, "std_wstring_return_std_wstring", "embind")
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
					size, err := obj.CallMethod(ctx, obj, "size")
					Expect(err).To(BeNil())
					Expect(size).To(Equal(uint32(10)))

					_, err = obj.CallMethod(ctx, obj, "resize", uint32(12), int32(1))
					Expect(err).To(BeNil())

					size, err = obj.CallMethod(ctx, obj, "size")
					Expect(err).To(BeNil())
					Expect(size).To(Equal(uint32(12)))

					val, err := obj.CallMethod(ctx, obj, "get", uint32(1))
					Expect(err).To(BeNil())
					Expect(val).To(Equal(int32(1)))

					_, err = obj.CallMethod(ctx, obj, "set", uint32(1), int32(2))
					Expect(err).To(BeNil())

					val, err = obj.CallMethod(ctx, obj, "get", uint32(1))
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
					size, err := obj.CallMethod(ctx, obj, "size")
					Expect(err).To(BeNil())
					Expect(size).To(Equal(uint32(1)))

					val, err := obj.CallMethod(ctx, obj, "get", int32(10))
					Expect(err).To(BeNil())
					Expect(val).To(Equal("This is a string."))

					val, err = obj.CallMethod(ctx, obj, "get", int32(1))
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
			Expect(constantsMap).To(HaveLen(15))
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
	})
})

var _ = Describe("Using embind structs", Label("library"), func() {
	When("using the structs", func() {
		It("can be created with an array as input", func() {
			res, err := engine.CallPublicSymbol(ctx, "findPersonAtLocation", []any{float32(1), float32(2)})
			Expect(err).To(BeNil())
			Expect(res).To(Equal(map[string]any{"name": "", "age": int32(12)}))
			// @todo: why doesn't it detect the array field?
		})
	})
})
