package tests

import (
	"context"
	"log"
	"os"
	"testing"

	embind_external "github.com/jerbob92/wazero-emscripten-embind"
	embind "github.com/jerbob92/wazero-emscripten-embind/internal"
	"github.com/jerbob92/wazero-emscripten-embind/tests/generated"
	"github.com/jerbob92/wazero-emscripten-embind/types"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/emscripten"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestEmbindGenerated(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Embind generated Suite")
}

var ctx = context.Background()

//var ctx = context.WithValue(context.Background(), experimental.FunctionListenerFactoryKey{}, logging.NewLoggingListenerFactory(os.Stdout))

var engine embind_external.Engine
var runtime wazero.Runtime
var mod api.Module
var wasmData []byte
var compiledModule wazero.CompiledModule

var _ = BeforeSuite(func() {
	wasm, err := os.ReadFile("../testdata/wasm/tests.wasm")
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

	compiledModule, err = runtime.CompileModule(ctx, wasm)
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
})

var _ = AfterSuite(func() {
	runtime.Close(ctx)
})

var _ = BeforeEach(func() {
	moduleConfig := wazero.NewModuleConfig().
		WithStartFunctions("_initialize").
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithName("")

	var err error
	engine = embind_external.CreateEngine(embind_external.NewConfig())
	ctx = engine.Attach(ctx)
	mod, err = runtime.InstantiateModule(ctx, compiledModule, moduleConfig)
	if err != nil {
		Expect(err).To(BeNil())
		return
	}

	err = generated.Attach(engine)
	if err != nil {
		Expect(err).To(BeNil())
		return
	}

	err = engine.SetDelayFunction(nil)
	Expect(err).To(BeNil())

	emvalHandleCount := engine.CountEmvalHandles()
	Expect(emvalHandleCount).To(Equal(0))
})

var _ = AfterEach(func() {
	err := engine.FlushPendingDeletes(ctx)
	Expect(err).To(BeNil())

	emvalHandleCount := engine.CountEmvalHandles()
	Expect(emvalHandleCount).To(Equal(0))

	mod.Close(ctx)
})

var _ = Describe("executing original embind tests", Label("library"), func() {
	When("access to base class members", func() {
		It("method name in derived class silently overrides inherited name", func() {
			derived, err := generated.NewClassDerived(engine, ctx)
			Expect(err).To(BeNil())
			Expect(derived.GetClassName(ctx)).To(Equal("Derived"))
			err = derived.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("can reference base method from derived class", func() {
			derived, err := generated.NewClassDerived(engine, ctx)
			Expect(err).To(BeNil())
			Expect(derived.GetClassNameFromBase(ctx)).To(Equal("Base"))
			err = derived.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("can reference base method from doubly derived class", func() {
			derivedTwice, err := generated.NewClassDerivedTwice(engine, ctx)
			Expect(err).To(BeNil())
			Expect(derivedTwice.GetClassNameFromBase(ctx)).To(Equal("Base"))
			err = derivedTwice.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("can reference base method through unbound classes", func() {
			derivedThrice, err := generated.NewClassDerivedThrice(engine, ctx)
			Expect(err).To(BeNil())
			Expect(derivedThrice.GetClassNameFromBase(ctx)).To(Equal("Base"))
			err = derivedThrice.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("property name in derived class hides identically named property in base class for set", func() {
			derived, err := generated.NewClassDerived(engine, ctx)
			Expect(err).To(BeNil())

			err = derived.SetMember(ctx, 7)
			Expect(err).To(BeNil())

			err = derived.SetPropertyMember(ctx, 17)
			Expect(err).To(BeNil())

			member, err := derived.GetMember(ctx)
			Expect(err).To(BeNil())
			Expect(member).To(Equal(int32(17)))

			err = derived.Delete(ctx)
			Expect(err).To(BeNil())
		})
		It("can reference base property from derived class for get", func() {
			derived, err := generated.NewClassDerived(engine, ctx)
			Expect(err).To(BeNil())

			err = derived.SetMember(ctx, 5)
			Expect(err).To(BeNil())

			member, err := derived.GetPropertyMember(ctx)
			Expect(err).To(BeNil())
			Expect(member).To(Equal(int32(5)))

			err = derived.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("can reference property of any base class for get when multiply derived", func() {
			derived, err := generated.NewClassMultiplyDerived(engine, ctx)
			Expect(err).To(BeNil())

			err = derived.SetMember(ctx, 11)
			Expect(err).To(BeNil())

			member, err := derived.GetPropertyMember(ctx)
			Expect(err).To(BeNil())
			Expect(member).To(Equal(int32(11)))

			err = derived.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("can reference base property from derived class for set", func() {
			derived, err := generated.NewClassDerived(engine, ctx)
			Expect(err).To(BeNil())

			err = derived.SetPropertyBaseMember(ctx, 32)
			Expect(err).To(BeNil())

			member, err := derived.GetBaseMember(ctx)
			Expect(err).To(BeNil())
			Expect(member).To(Equal(int32(32)))

			err = derived.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("can reference property of any base for set when multiply derived", func() {
			derived, err := generated.NewClassMultiplyDerived(engine, ctx)
			Expect(err).To(BeNil())

			err = derived.SetPropertyBaseMember(ctx, 32)
			Expect(err).To(BeNil())

			member, err := derived.GetBaseMember(ctx)
			Expect(err).To(BeNil())
			Expect(member).To(Equal(int32(32)))

			err = derived.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("can reach around derived property to access base property with same name for get", func() {
			derived, err := generated.NewClassDerived(engine, ctx)
			Expect(err).To(BeNil())

			err = derived.SetMember(ctx, 12)
			Expect(err).To(BeNil())

			err = derived.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("if deriving from second base adjusts pointer", func() {
			derived, err := generated.HasTwoBases(engine, ctx)
			Expect(err).To(BeNil())

			derivedClass := derived.(*generated.ClassHasTwoBases)

			getField, err := derivedClass.GetField(ctx)
			Expect(getField).To(Equal("Base2"))

			err = derivedClass.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("properties adjust pointer", func() {
			derived, err := generated.HasTwoBases(engine, ctx)
			Expect(err).To(BeNil())

			derivedClass := derived.(*generated.ClassHasTwoBases)

			err = derivedClass.SetPropertyField(ctx, "Foo")
			Expect(err).To(BeNil())

			getField, err := derivedClass.GetField(ctx)
			Expect(getField).To(Equal("Foo"))

			getFieldProperty, err := derivedClass.GetPropertyField(ctx)
			Expect(getFieldProperty).To(Equal("Foo"))

			err = derivedClass.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("class functions are inherited in subclasses", func() {
			classFunction, err := generated.ClassBaseStaticClassFunction(engine, ctx)
			Expect(err).To(BeNil())
			Expect(classFunction).To(Equal("Base"))

			classFunction, err = generated.ClassDerivedStaticClassFunction(engine, ctx)
			Expect(err).To(BeNil())
			Expect(classFunction).To(Equal("Derived"))

			classFunction, err = generated.ClassDerivedTwiceStaticClassFunction(engine, ctx)
			Expect(err).To(BeNil())
			Expect(classFunction).To(Equal("Derived"))
		})

		// @todo: do we want to test this? You can't really reach this in Go anyway?
		/*
		   test("calling method on unrelated class throws error", function() {
		       var a = new cm.HasTwoBases;
		       var e = assert.throws(cm.BindingError, function() {
		           cm.Derived.prototype.setMember.call(a, "foo");
		       });
		       assert.equal('Expected null or instance of Derived, got an instance of Base2', e.message);
		       a.delete();

		       // Base1 and Base2 both have the method 'getField()' exposed - make sure
		       // that calling the Base2 function with a 'this' instance of Base1 doesn't accidentally work!
		       var b = new cm.Base1;
		       var e = assert.throws(cm.BindingError, function() {
		           cm.Base2.prototype.getField.call(b);
		       });
		       assert.equal('Expected null or instance of Base2, got an instance of Base1', e.message);
		       b.delete();
		   });

		   test("calling method with invalid this throws error", function() {
		       var e = assert.throws(cm.BindingError, function() {
		           cm.Derived.prototype.setMember.call(undefined, "foo");
		       });
		       assert.equal('Cannot pass "[object global]" as a Derived*', e.message);

		       var e = assert.throws(cm.BindingError, function() {
		           cm.Derived.prototype.setMember.call(true, "foo");
		       });
		       assert.equal('Cannot pass "true" as a Derived*', e.message);

		       var e = assert.throws(cm.BindingError, function() {
		           cm.Derived.prototype.setMember.call(null, "foo");
		       });
		       assert.equal('Cannot pass "[object global]" as a Derived*', e.message);

		       var e = assert.throws(cm.BindingError, function() {
		           cm.Derived.prototype.setMember.call(42, "foo");
		       });
		       assert.equal('Cannot pass "42" as a Derived*', e.message);

		       var e = assert.throws(cm.BindingError, function() {
		           cm.Derived.prototype.setMember.call("this", "foo");
		       });
		       assert.equal('Cannot pass "this" as a Derived*', e.message);

		       var e = assert.throws(cm.BindingError, function() {
		           cm.Derived.prototype.setMember.call({}, "foo");
		       });
		       assert.equal('Cannot pass "[object Object]" as a Derived*', e.message);
		   });

		   test("setting and getting property on unrelated class throws error", function() {
		       var a = new cm.HasTwoBases;
		       var e = assert.throws(cm.BindingError, function() {
		           Object.getOwnPropertyDescriptor(cm.HeldBySmartPtr.prototype, 'i').set.call(a, 10);
		       });
		       assert.equal('HeldBySmartPtr.i setter incompatible with "this" of type HasTwoBases', e.message);

		       var e = assert.throws(cm.BindingError, function() {
		           Object.getOwnPropertyDescriptor(cm.HeldBySmartPtr.prototype, 'i').get.call(a);
		       });
		       assert.equal('HeldBySmartPtr.i getter incompatible with "this" of type HasTwoBases', e.message);

		       a.delete();
		   });
		*/

	})

	When("automatic upcasting of parameters passed to C++", func() {
		It("raw pointer argument is upcast to parameter type", func() {
			derived, err := generated.NewClassDerived(engine, ctx)
			Expect(err).To(BeNil())

			name, err := generated.Embind_test_get_class_name_via_base_ptr(engine, ctx, derived)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("Base"))
			err = derived.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("automatic raw pointer upcasting works with multiple inheritance", func() {
			derived, err := generated.NewClassMultiplyDerived(engine, ctx)
			Expect(err).To(BeNil())

			name, err := generated.Embind_test_get_class_name_via_base_ptr(engine, ctx, derived)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("Base"))
			err = derived.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("automatic raw pointer upcasting does not change local pointer", func() {
			derived, err := generated.NewClassMultiplyDerived(engine, ctx)
			Expect(err).To(BeNil())

			_, err = generated.Embind_test_get_class_name_via_base_ptr(engine, ctx, derived)
			Expect(err).To(BeNil())

			name, err := derived.GetClassName(ctx)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("MultiplyDerived"))

			err = derived.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("passing incompatible raw pointer to method throws exception", func() {
			base, err := generated.NewClassBase(engine, ctx)
			Expect(err).To(BeNil())

			_, err = generated.Embind_test_get_class_name_via_second_base_ptr(engine, ctx, base)
			Expect(err).To(Not(BeNil()))
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("expected null or instance of SecondBase, got an instance of Base"))
			}

			err = base.Delete(ctx)
			Expect(err).To(BeNil())

		})

		// raw polymorphic
		It("polymorphic raw pointer argument is upcast to parameter type", func() {
			derived, err := generated.NewClassPolyDerived(engine, ctx)
			Expect(err).To(BeNil())

			name, err := generated.Embind_test_get_class_name_via_polymorphic_base_ptr(engine, ctx, derived)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolyBase"))

			err = derived.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("automatic polymorphic raw pointer upcasting works with multiple inheritance", func() {
			derived, err := generated.NewClassPolyMultiplyDerived(engine, ctx)
			Expect(err).To(BeNil())

			name, err := generated.Embind_test_get_class_name_via_polymorphic_base_ptr(engine, ctx, derived)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolyBase"))

			err = derived.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("passing incompatible raw polymorphic pointer to method throws exception", func() {
			base, err := generated.NewClassPolyBase(engine, ctx)
			Expect(err).To(BeNil())

			_, err = generated.Embind_test_get_class_name_via_polymorphic_second_base_ptr(engine, ctx, base)
			Expect(err).To(Not(BeNil()))
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("expected null or instance of PolySecondBase, got an instance of PolyBase"))
			}

			err = base.Delete(ctx)
			Expect(err).To(BeNil())
		})

		// smart
		It("can pass smart pointer to raw pointer parameter", func() {
			smartBase, err := generated.Embind_test_return_smart_base_ptr(engine, ctx)
			Expect(err).To(BeNil())

			name, err := generated.Embind_test_get_class_name_via_base_ptr(engine, ctx, smartBase)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("Base"))

			err = smartBase.(*generated.ClassBase).Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("can pass and upcast smart pointer to raw pointer parameter", func() {
			smartDerived, err := generated.Embind_test_return_smart_derived_ptr(engine, ctx)
			Expect(err).To(BeNil())

			name, err := generated.Embind_test_get_class_name_via_base_ptr(engine, ctx, smartDerived)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("Base"))

			err = smartDerived.(*generated.ClassDerived).Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("smart pointer argument is upcast to parameter type", func() {
			derived, err := generated.Embind_test_return_smart_derived_ptr(engine, ctx)
			Expect(err).To(BeNil())

			// Todo: can we implement these?
			//assert.instanceof(derived, cm.Derived)
			//assert.instanceof(derived, cm.Base)

			name, err := generated.Embind_test_get_class_name_via_smart_base_ptr(engine, ctx, derived)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("Base"))

			err = derived.(*generated.ClassDerived).Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("return smart derived ptr as base", func() {
			derived, err := generated.Embind_test_return_smart_derived_ptr_as_base(engine, ctx)
			Expect(err).To(BeNil())

			name, err := generated.Embind_test_get_virtual_class_name_via_smart_polymorphic_base_ptr(engine, ctx, derived)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolyDerived"))

			name, err = derived.(*generated.ClassPolyDerived).GetClassName(ctx)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolyDerived"))

			err = derived.(*generated.ClassPolyDerived).Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("return smart derived ptr as val", func() {
			derived, err := generated.Embind_test_return_smart_derived_ptr_as_val(engine, ctx)
			Expect(err).To(BeNil())

			name, err := generated.Embind_test_get_virtual_class_name_via_smart_polymorphic_base_ptr(engine, ctx, derived.(*generated.ClassPolyDerived))
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolyDerived"))

			err = derived.(*generated.ClassPolyDerived).Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("automatic smart pointer upcasting works with multiple inheritance", func() {
			derived, err := generated.Embind_test_return_smart_multiply_derived_ptr(engine, ctx)
			Expect(err).To(BeNil())

			name, err := generated.Embind_test_get_class_name_via_smart_base_ptr(engine, ctx, derived)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("Base"))

			err = derived.(*generated.ClassMultiplyDerived).Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("automatically upcasted smart pointer parameter shares ownership with original argument", func() {
			derived, err := generated.Embind_test_return_smart_multiply_derived_ptr(engine, ctx)
			Expect(err).To(BeNil())

			instanceCount, err := derived.(*generated.ClassMultiplyDerived).StaticGetInstanceCount(ctx)
			Expect(err).To(BeNil())
			Expect(instanceCount).To(Equal(int32(1)))

			err = generated.Embind_save_smart_base_pointer(engine, ctx, derived)
			Expect(err).To(BeNil())

			instanceCount, err = derived.(*generated.ClassMultiplyDerived).StaticGetInstanceCount(ctx)
			Expect(err).To(BeNil())
			Expect(instanceCount).To(Equal(int32(1)))

			err = derived.(*generated.ClassMultiplyDerived).Delete(ctx)
			Expect(err).To(BeNil())

			instanceCount, err = derived.(*generated.ClassMultiplyDerived).StaticGetInstanceCount(ctx)
			Expect(err).To(BeNil())
			Expect(instanceCount).To(Equal(int32(1)))

			err = generated.Embind_save_smart_base_pointer(engine, ctx, nil)
			Expect(err).To(BeNil())

			instanceCount, err = derived.(*generated.ClassMultiplyDerived).StaticGetInstanceCount(ctx)
			Expect(err).To(BeNil())
			Expect(instanceCount).To(Equal(int32(0)))
		})

		// smart polymorphic
		It("smart polymorphic pointer argument is upcast to parameter type", func() {
			derived, err := generated.Embind_test_return_smart_polymorphic_derived_ptr(engine, ctx)
			Expect(err).To(BeNil())

			name, err := generated.Embind_test_get_class_name_via_smart_polymorphic_base_ptr(engine, ctx, derived)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolyBase"))

			err = derived.(*generated.ClassPolyDerived).Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("automatic smart polymorphic pointer upcasting works with multiple inheritance", func() {
			derived, err := generated.Embind_test_return_smart_polymorphic_multiply_derived_ptr(engine, ctx)
			Expect(err).To(BeNil())

			name, err := generated.Embind_test_get_class_name_via_smart_polymorphic_base_ptr(engine, ctx, derived)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolyBase"))

			err = derived.(*generated.ClassPolyMultiplyDerived).Delete(ctx)
			Expect(err).To(BeNil())
		})

	})
	When("automatic downcasting of return values received from C++", func() {
		// raw
		It("non-polymorphic raw pointers are not downcast and do not break automatic casting mechanism", func() {
			base, err := generated.Embind_test_return_raw_derived_ptr_as_base(engine, ctx)
			Expect(err).To(BeNil())

			name, err := base.(*generated.ClassBase).GetClassName(ctx)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("Base"))

			err = base.(*generated.ClassBase).Delete(ctx)
			Expect(err).To(BeNil())
		})

		// raw polymorphic
		It("polymorphic raw pointer return value is downcast to allocated type (if that is bound)", func() {
			derived, err := generated.Embind_test_return_raw_polymorphic_derived_ptr_as_base(engine, ctx)
			Expect(err).To(BeNil())

			name, err := derived.(*generated.ClassPolyDerived).GetClassName(ctx)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolyDerived"))

			//assert.instanceof(derived, cm.PolyBase)
			//assert.instanceof(derived, cm.PolyDerived)

			siblingDerived, err := generated.Embind_test_return_raw_polymorphic_sibling_derived_ptr_as_base(engine, ctx)
			Expect(err).To(BeNil())

			name, err = siblingDerived.(*generated.ClassPolySiblingDerived).GetClassName(ctx)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolySiblingDerived"))

			err = siblingDerived.(*generated.ClassPolySiblingDerived).Delete(ctx)
			Expect(err).To(BeNil())

			err = derived.(*generated.ClassPolyDerived).Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("polymorphic raw pointer return value is downcast to the most derived bound type", func() {
			derivedThrice, err := generated.Embind_test_return_raw_polymorphic_derived_four_times_not_bound_as_base(engine, ctx)
			Expect(err).To(BeNil())

			// if the actual returned type is not bound, then don't assume anything
			name, err := derivedThrice.(*generated.ClassPolyBase).GetClassName(ctx)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolyBase"))

			// if we ever fix this, then reverse the assertion (comment from Emscripten)
			//assert.equal("PolyDerivedThrice", derivedThrice.getClassName());

			err = derivedThrice.(*generated.ClassPolyBase).Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("polymorphic smart pointer return value is downcast to the most derived type which has an associated smart pointer", func() {
			derived, err := generated.Embind_test_return_poly_derived_twice_without_smart_pointer_as_poly_base(engine, ctx)
			Expect(err).To(BeNil())

			// if the actual returned type is not bound, then don't assume anything
			name, err := derived.(*generated.ClassPolyBase).GetClassName(ctx)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolyBase"))

			// if we ever fix this, then reverse the assertion (comment from Emscripten)
			//assert.equal("PolyDerived", derivedThrice.getClassName());

			err = derived.(*generated.ClassPolyBase).Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("automatic downcasting works with multiple inheritance", func() {
			base, err := generated.Embind_test_return_raw_polymorphic_multiply_derived_ptr_as_base(engine, ctx)
			Expect(err).To(BeNil())

			secondBase, err := generated.Embind_test_return_raw_polymorphic_multiply_derived_ptr_as_second_base(engine, ctx)
			Expect(err).To(BeNil())

			name, err := base.(*generated.ClassPolyMultiplyDerived).GetClassName(ctx)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolyMultiplyDerived"))

			// embind does not support multiple inheritance
			//assert.equal("PolyMultiplyDerived", secondBase.getClassName());

			err = secondBase.(*generated.ClassPolySecondBase).Delete(ctx)
			Expect(err).To(BeNil())

			err = base.(*generated.ClassPolyMultiplyDerived).Delete(ctx)
			Expect(err).To(BeNil())
		})

		// smart polymorphic
		It("automatically downcasting a smart pointer does not change the underlying pointer", func() {
			err := generated.ClassPolyDerivedStaticSetPtrDerived(engine, ctx)
			Expect(err).To(BeNil())

			name, err := generated.ClassPolyDerivedStaticGetPtrClassName(engine, ctx)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolyBase"))

			derived, err := generated.ClassPolyDerivedStaticGetPtr(engine, ctx)
			Expect(err).To(BeNil())

			name, err = derived.(*generated.ClassPolyDerived).GetClassName(ctx)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolyDerived"))

			name, err = generated.ClassPolyDerivedStaticGetPtrClassName(engine, ctx)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolyBase"))

			err = derived.(*generated.ClassPolyDerived).Delete(ctx)
			Expect(err).To(BeNil())

			err = generated.ClassPolyDerivedStaticReleasePtr(engine, ctx)
			Expect(err).To(BeNil())
		})

		It("polymorphic smart pointer return value is actual allocated type (when bound)", func() {
			derived, err := generated.Embind_test_return_smart_polymorphic_derived_ptr_as_base(engine, ctx)
			Expect(err).To(BeNil())

			name, err := derived.(*generated.ClassPolyDerived).GetClassName(ctx)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolyDerived"))

			siblingDerived, err := generated.Embind_test_return_smart_polymorphic_sibling_derived_ptr_as_base(engine, ctx)
			Expect(err).To(BeNil())

			name, err = siblingDerived.(*generated.ClassPolySiblingDerived).GetClassName(ctx)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("PolySiblingDerived"))

			err = siblingDerived.(*generated.ClassPolySiblingDerived).Delete(ctx)
			Expect(err).To(BeNil())

			err = derived.(*generated.ClassPolyDerived).Delete(ctx)
			Expect(err).To(BeNil())
		})

	})
	When("string", func() {
		// var stdStringIsUTF8 = (cm.getCompilerSetting('EMBIND_STD_STRING_IS_UTF8') === true);
		// We only support that this is true.
		stdStringIsUTF8 := true

		It("non-ascii strings", func() {

			//if(stdStringIsUTF8) {

			//ASCII
			expected := "aei"
			//Latin-1 Supplement
			expected += "\u00E1\u00E9\u00ED"
			//Greek
			expected += "\u03B1\u03B5\u03B9"
			//Cyrillic
			expected += "\u0416\u041B\u0424"
			//CJK
			expected += "\u5F9E\u7345\u5B50"
			//Euro sign
			expected += "\u20AC"

			//} else {
			//	for (var i = 0; i < 128; ++i) {
			//			expected += String.fromCharCode(128 + i);
			//		}
			//	}

			non_ascii_string, err := generated.Get_non_ascii_string(engine, ctx, stdStringIsUTF8)
			Expect(err).To(BeNil())
			Expect(non_ascii_string).To(Equal(expected))
		})

		/*
		   if(!stdStringIsUTF8) {
		       It("passing non-8-bit strings from JS to std::string throws", func() {
		           assert.throws(cm.BindingError, function() {
		               cm.emval_test_take_and_return_std_string("\u1234");
		           });
		       });
		   }
		*/

		It("can't pass integers as strings", func() {
			_, err := engine.CallPublicSymbol(ctx, "emval_test_take_and_return_std_string", 10)
			Expect(err).To(Not(BeNil()))
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("value must be of type string"))
			}
		})

		/*
					   It("can pass Uint8Array to std::string", func() {
					       var e = cm.emval_test_take_and_return_std_string(new Uint8Array([65, 66, 67, 68]));
					       assert.equal('ABCD', e);
					   });

					It("can pass Uint8ClampedArray to std::string", func() {
					       var e = cm.emval_test_take_and_return_std_string(new Uint8ClampedArray([65, 66, 67, 68]));
					       assert.equal('ABCD', e);
					   });

					It("can pass Int8Array to std::string", func() {
					       var e = cm.emval_test_take_and_return_std_string(new Int8Array([65, 66, 67, 68]));
					       assert.equal('ABCD', e);
					   });

					It("can pass ArrayBuffer to std::string", func() {
					       var e = cm.emval_test_take_and_return_std_string((new Int8Array([65, 66, 67, 68])).buffer);
					       assert.equal('ABCD', e);
					   });

					It("can pass Uint8Array to std::basic_string<unsigned char>", func() {
					       var e = cm.emval_test_take_and_return_std_basic_string_unsigned_char(new Uint8Array([65, 66, 67, 68]));
					       assert.equal('ABCD', e);
					   });

					It("can pass long string to std::basic_string<unsigned char>", func() {
					       var s = 'this string is long enough to exceed the short string optimization';
					       var e = cm.emval_test_take_and_return_std_basic_string_unsigned_char(s);
					       assert.equal(s, e);
					   });

					It("can pass Uint8ClampedArray to std::basic_string<unsigned char>", func() {
					       var e = cm.emval_test_take_and_return_std_basic_string_unsigned_char(new Uint8ClampedArray([65, 66, 67, 68]));
					       assert.equal('ABCD', e);
					   });


					It("can pass Int8Array to std::basic_string<unsigned char>", func() {
					       var e = cm.emval_test_take_and_return_std_basic_string_unsigned_char(new Int8Array([65, 66, 67, 68]));
					       assert.equal('ABCD', e);
					   });

					It("can pass ArrayBuffer to std::basic_string<unsigned char>", func() {
					       var e = cm.emval_test_take_and_return_std_basic_string_unsigned_char((new Int8Array([65, 66, 67, 68])).buffer);
					       assert.equal('ABCD', e);
					   });

					It("can pass string to std::string", func() {
					       //var string = stdStringIsUTF8?"aeiáéíαειЖЛФ從獅子€":"ABCD";

							string := "aeiáéíαειЖЛФ從獅子€"

					       var e = cm.emval_test_take_and_return_std_string(string);
					       assert.equal(string, e);
					   });

					   var utf16TestString = String.fromCharCode(10) +
					       String.fromCharCode(1234) +
					       String.fromCharCode(2345) +
					       String.fromCharCode(65535);

					   var utf32TestString = String.fromCharCode(10) +
					       String.fromCharCode(1234) +
					       String.fromCharCode(2345) +
					       String.fromCharCode(55357) +
					       String.fromCharCode(56833) +
					       String.fromCharCode(55357) +
					       String.fromCharCode(56960);

			It("non-ascii wstrings", func() {
					       assert.equal(utf16TestString, cm.get_non_ascii_wstring());
					   });

			It("non-ascii u16strings", func() {
					       assert.equal(utf16TestString, cm.get_non_ascii_u16string());
					   });

			It("non-ascii u32strings", func() {
					       assert.equal(utf32TestString, cm.get_non_ascii_u32string());
					   });

			It("passing unicode (wide) string into C++", func() {
					       assert.equal(utf16TestString, cm.take_and_return_std_wstring(utf16TestString));
					   });

			It("passing unicode (utf-16) string into C++", func() {
					       assert.equal(utf16TestString, cm.take_and_return_std_u16string(utf16TestString));
					   });

			It("passing unicode (utf-32) string into C++", func() {
					       assert.equal(utf32TestString, cm.take_and_return_std_u32string(utf32TestString));
					   });

					   //if (cm.isMemoryGrowthEnabled) {
			It("can access a literal wstring after a memory growth", func() {
					           cm.force_memory_growth();
					           assert.equal("get_literal_wstring", cm.get_literal_wstring());
					       });

			It("can access a literal u16string after a memory growth", func() {
					           cm.force_memory_growth();
					           assert.equal("get_literal_u16string", cm.get_literal_u16string());
					       });

			It("can access a literal u32string after a memory growth", func() {
					           cm.force_memory_growth();
					           assert.equal("get_literal_u32string", cm.get_literal_u32string());
					       });
					   //}
		*/
	})
	When("embind", func() {
		It("value creation", func() {
			newInteger, err := generated.Emval_test_new_integer(engine, ctx)
			Expect(err).To(BeNil())
			Expect(newInteger).To(Equal(int32(15)))

			newString, err := generated.Emval_test_new_string(engine, ctx)
			Expect(err).To(BeNil())
			Expect(newString).To(Equal("Hello everyone"))

			newStringFromVal, err := generated.Emval_test_get_string_from_val(engine, ctx, map[string]any{"key": "Hello everyone"})
			Expect(err).To(BeNil())
			Expect(newStringFromVal).To(Equal("Hello everyone"))

			object, err := generated.Emval_test_new_object(engine, ctx)
			Expect(err).To(BeNil())
			Expect(object).To(HaveKeyWithValue("foo", "bar"))
			Expect(object).To(HaveKeyWithValue("baz", int32(1)))
		})

		It("pass const reference to primitive", func() {
			const_ref_adder, err := generated.Const_ref_adder(engine, ctx, 1, 2)
			Expect(err).To(BeNil())
			Expect(const_ref_adder).To(Equal(float32(3)))
		})

		It("get instance pointer as value", func() {
			v, err := generated.Emval_test_instance_pointer(engine, ctx)
			Expect(err).To(BeNil())
			_, ok := v.(*generated.ClassDummyForPointer)
			Expect(ok).To(BeTrue())
		})

		It("cast value to instance pointer using as<T*>", func() {
			v, err := generated.Emval_test_instance_pointer(engine, ctx)
			Expect(err).To(BeNil())
			p_value, err := generated.Emval_test_value_from_instance_pointer(engine, ctx, v)
			Expect(err).To(BeNil())

			Expect(p_value).To(Equal(int32(42)))
		})

		It("passthrough", func() {
			a := map[string]any{"foo": "bar"}
			b, err := generated.Emval_test_passthrough(engine, ctx, a)
			Expect(err).To(BeNil())

			a["bar"] = "baz"
			Expect(b.(map[string]any)["bar"]).To(Equal("baz"))

			emvalhandles := engine.CountEmvalHandles()
			Expect(emvalhandles).To(Equal(0))
		})

		It("void return converts to undefined", func() {
			err := generated.Emval_test_return_void(engine, ctx)
			Expect(err).To(BeNil())

			val, err := engine.CallPublicSymbol(ctx, "emval_test_return_void")
			Expect(err).To(BeNil())
			Expect(val).To(BeNil())
		})

		It("booleans can be marshalled", func() {
			not, err := generated.Emval_test_not(engine, ctx, true)
			Expect(err).To(BeNil())
			Expect(not).To(BeFalse())

			not, err = generated.Emval_test_not(engine, ctx, false)
			Expect(err).To(BeNil())
			Expect(not).To(BeTrue())
		})

		It("val.is_undefined() is functional", func() {
			emval_test_is_undefined, err := generated.Emval_test_is_undefined(engine, ctx, types.Undefined)
			Expect(err).To(BeNil())
			Expect(emval_test_is_undefined).To(BeTrue())

			emval_test_is_undefined, err = generated.Emval_test_is_undefined(engine, ctx, true)
			Expect(err).To(BeNil())
			Expect(emval_test_is_undefined).To(BeFalse())

			emval_test_is_undefined, err = generated.Emval_test_is_undefined(engine, ctx, false)
			Expect(err).To(BeNil())
			Expect(emval_test_is_undefined).To(BeFalse())

			emval_test_is_undefined, err = generated.Emval_test_is_undefined(engine, ctx, nil)
			Expect(err).To(BeNil())
			Expect(emval_test_is_undefined).To(BeFalse())

			emval_test_is_undefined, err = generated.Emval_test_is_undefined(engine, ctx, struct{}{})
			Expect(err).To(BeNil())
			Expect(emval_test_is_undefined).To(BeFalse())
		})

		It("val.is_null() is functional", func() {
			emval_test_is_null, err := generated.Emval_test_is_null(engine, ctx, nil)
			Expect(err).To(BeNil())
			Expect(emval_test_is_null).To(BeTrue())

			emval_test_is_null, err = generated.Emval_test_is_null(engine, ctx, true)
			Expect(err).To(BeNil())
			Expect(emval_test_is_null).To(BeFalse())

			emval_test_is_null, err = generated.Emval_test_is_null(engine, ctx, false)
			Expect(err).To(BeNil())
			Expect(emval_test_is_null).To(BeFalse())

			emval_test_is_null, err = generated.Emval_test_is_null(engine, ctx, types.Undefined)
			Expect(err).To(BeNil())
			Expect(emval_test_is_null).To(BeFalse())

			emval_test_is_null, err = generated.Emval_test_is_null(engine, ctx, struct{}{})
			Expect(err).To(BeNil())
			Expect(emval_test_is_null).To(BeFalse())
		})

		It("val.is_true() is functional", func() {
			emval_test_is_true, err := generated.Emval_test_is_true(engine, ctx, true)
			Expect(err).To(BeNil())
			Expect(emval_test_is_true).To(BeTrue())

			emval_test_is_true, err = generated.Emval_test_is_true(engine, ctx, false)
			Expect(err).To(BeNil())
			Expect(emval_test_is_true).To(BeFalse())

			emval_test_is_true, err = generated.Emval_test_is_true(engine, ctx, nil)
			Expect(err).To(BeNil())
			Expect(emval_test_is_true).To(BeFalse())

			emval_test_is_true, err = generated.Emval_test_is_true(engine, ctx, types.Undefined)
			Expect(err).To(BeNil())
			Expect(emval_test_is_true).To(BeFalse())

			emval_test_is_true, err = generated.Emval_test_is_true(engine, ctx, struct{}{})
			Expect(err).To(BeNil())
			Expect(emval_test_is_true).To(BeFalse())
		})

		It("val.is_false() is functional", func() {
			emval_test_is_false, err := generated.Emval_test_is_false(engine, ctx, false)
			Expect(err).To(BeNil())
			Expect(emval_test_is_false).To(BeTrue())

			emval_test_is_false, err = generated.Emval_test_is_false(engine, ctx, true)
			Expect(err).To(BeNil())
			Expect(emval_test_is_false).To(BeFalse())

			emval_test_is_false, err = generated.Emval_test_is_false(engine, ctx, nil)
			Expect(err).To(BeNil())
			Expect(emval_test_is_false).To(BeFalse())

			emval_test_is_false, err = generated.Emval_test_is_false(engine, ctx, types.Undefined)
			Expect(err).To(BeNil())
			Expect(emval_test_is_false).To(BeFalse())

			emval_test_is_false, err = generated.Emval_test_is_false(engine, ctx, struct{}{})
			Expect(err).To(BeNil())
			Expect(emval_test_is_false).To(BeFalse())
		})

		It("val.equals() is functional", func() {
			vals := []any{types.Undefined, nil, true, false, struct{}{}}
			for i := range vals {
				first := vals[i]
				for j := range vals {
					second := vals[j]
					isEqual, err := generated.Emval_test_equals(engine, ctx, first, second)
					Expect(err).To(BeNil())
					if i == j {
						Expect(isEqual).To(BeTrue())
					} else {
						Expect(isEqual).To(BeFalse())
					}
				}
			}
		})

		It("val.strictlyEquals() is functional", func() {
			vals := []any{types.Undefined, nil, true, false, struct{}{}}
			for i := range vals {
				first := vals[i]
				for j := range vals {
					second := vals[j]
					isEqual, err := generated.Emval_test_strictly_equals(engine, ctx, first, second)
					Expect(err).To(BeNil())
					if i == j {
						Expect(isEqual).To(BeTrue())
					} else {
						Expect(isEqual).To(BeFalse())
					}
				}
			}
		})

		/*
			test("passing Symbol or BigInt as floats always throws", function() {
			assert.throws(TypeError, function() { cm.const_ref_adder(Symbol('0'), 1); });
			assert.throws(TypeError, function() { cm.const_ref_adder(0n, 1); });
			});

			if (cm.getCompilerSetting('ASSERTIONS')) {
			test("can pass only number and boolean as floats with assertions", function() {
			assert.throws(TypeError, function() { cm.const_ref_adder(1, undefined); });
			assert.throws(TypeError, function() { cm.const_ref_adder(1, null); });
			assert.throws(TypeError, function() { cm.const_ref_adder(1, '2'); });
			});
			} else {
			test("can pass other types as floats without assertions", function() {
			assert.equal(3, cm.const_ref_adder(1, '2'));
			assert.equal(1, cm.const_ref_adder(1, null));  // null => 0
			assert.true(isNaN(cm.const_ref_adder(1, 'cannot parse')));
			assert.true(isNaN(cm.const_ref_adder(1, undefined)));  // undefined => NaN
			});
			}

			test("convert double to unsigned", function() {
			var rv = cm.emval_test_as_unsigned(1.5);
			assert.equal('number', typeof rv);
			assert.equal(1, rv);
			assert.equal(0, cm.count_emval_handles());
			});

			test("get length of array", function() {
			assert.equal(10, cm.emval_test_get_length([0, 1, 2, 3, 4, 5, 'a', 'b', 'c', 'd']));
			assert.equal(0, cm.count_emval_handles());
			});

			test("add a bunch of things", function() {
			assert.equal(66.0, cm.emval_test_add(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11));
			assert.equal(0, cm.count_emval_handles());
			});

			test("sum array", function() {
			assert.equal(66, cm.emval_test_sum([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]));
			assert.equal(0, cm.count_emval_handles());
			});

			test("strings", function() {
			assert.equal("foobar", "foo" + "bar");
			assert.equal("foobar", cm.emval_test_take_and_return_std_string("foobar"));

			assert.equal("foobar", cm.emval_test_take_and_return_std_string_const_ref("foobar"));
			});

			test("nuls pass through strings", function() {
			assert.equal("foo\0bar", cm.emval_test_take_and_return_std_string("foo\0bar"));
			});

			test("no memory leak when passing strings in by const reference", function() {
			cm.emval_test_take_and_return_std_string_const_ref("foobar");
			});

			test("can get global", function(){
			assert.equal((new Function("return this;"))(), cm.embind_test_getglobal());
			});

			test("can create new object", function() {
			assert.deepEqual({}, cm.embind_test_new_Object());
			});

			test("can invoke constructors with arguments", function() {
			function constructor(i, s, argument) {
			this.i = i;
			this.s = s;
			this.argument = argument;
			}
			constructor.prototype.method = function() {
			return this.argument;
			};
			var x = {};
			var instance = cm.embind_test_new_factory(constructor, x);
			assert.equal(10, instance.i);
			assert.equal("hello", instance.s);
			assert.equal(x, instance.argument);
			});

			test("can return module property objects", function() {
			assert.equal(cm.HEAP8, cm.get_module_property("HEAP8"));
			});

			test("can return big class instances", function() {
			var c = cm.embind_test_return_big_class_instance();
			assert.equal(11, c.member);
			c.delete();
			});

			test("can return small class instances", function() {
			var c = cm.embind_test_return_small_class_instance();
			assert.equal(7, c.member);
			c.delete();
			});

			test("can pass small class instances", function() {
			var c = new cm.SmallClass();
			var m = cm.embind_test_accept_small_class_instance(c);
			assert.equal(7, m);
			c.delete();
			});

			test("can pass big class instances", function() {
			var c = new cm.BigClass();
			var m = cm.embind_test_accept_big_class_instance(c);
			assert.equal(11, m);
			c.delete();
			});

			test("can pass unique_ptr", function() {
			var p = cm.embind_test_return_unique_ptr(42);
			var m = cm.embind_test_accept_unique_ptr(p);
			assert.equal(42, m);
			});

			test("can pass unique_ptr to constructor", function() {
			var c = new cm.embind_test_construct_class_with_unique_ptr(42);
			assert.equal(42, c.getValue());
			c.delete();
			});

			test("can get member classes then call its member functions", function() {
			var p = new cm.ParentClass();
			var c = p.getBigClass();
			var m = c.getMember();
			assert.equal(11, m);
			c.delete();
			p.delete();
			});

			test('C++ -> JS primitive type range checks', function() {
			// all types should have zero.
			assert.equal("0", cm.char_to_string(0));
			assert.equal("0", cm.signed_char_to_string(0));
			assert.equal("0", cm.unsigned_char_to_string(0));
			assert.equal("0", cm.short_to_string(0));
			assert.equal("0", cm.unsigned_short_to_string(0));
			assert.equal("0", cm.int_to_string(0));
			assert.equal("0", cm.unsigned_int_to_string(0));
			assert.equal("0", cm.long_to_string(0));
			assert.equal("0", cm.unsigned_long_to_string(0));

			// all types should have positive values.
			assert.equal("5", cm.char_to_string(5));
			assert.equal("5", cm.signed_char_to_string(5));
			assert.equal("5", cm.unsigned_char_to_string(5));
			assert.equal("5", cm.short_to_string(5));
			assert.equal("5", cm.unsigned_short_to_string(5));
			assert.equal("5", cm.int_to_string(5));
			assert.equal("5", cm.unsigned_int_to_string(5));
			assert.equal("5", cm.long_to_string(5));
			assert.equal("5", cm.unsigned_long_to_string(5));

			// signed types should have negative values.
			assert.equal("-5", cm.char_to_string(-5)); // Assuming char as signed.
			assert.equal("-5", cm.signed_char_to_string(-5));
			assert.equal("-5", cm.short_to_string(-5));
			assert.equal("-5", cm.int_to_string(-5));
			assert.equal("-5", cm.long_to_string(-5));

			// assumptions: char == signed char == 8 bits
			//              unsigned char == 8 bits
			//              short == 16 bits
			//              int == long == 32 bits

			// all types should have their max positive values.
			assert.equal("127", cm.char_to_string(127));
			assert.equal("127", cm.signed_char_to_string(127));
			assert.equal("255", cm.unsigned_char_to_string(255));
			assert.equal("32767", cm.short_to_string(32767));
			assert.equal("65535", cm.unsigned_short_to_string(65535));
			assert.equal("2147483647", cm.int_to_string(2147483647));
			assert.equal("4294967295", cm.unsigned_int_to_string(4294967295));
			assert.equal("2147483647", cm.long_to_string(2147483647));
			assert.equal("4294967295", cm.unsigned_long_to_string(4294967295));

			// signed types should have their min negative values.
			assert.equal("-128", cm.char_to_string(-128));
			assert.equal("-128", cm.signed_char_to_string(-128));
			assert.equal("-32768", cm.short_to_string(-32768));
			assert.equal("-2147483648", cm.int_to_string(-2147483648));
			assert.equal("-2147483648", cm.long_to_string(-2147483648));

			// passing out of range values should fail with assertions.
			if (cm.getCompilerSetting('ASSERTIONS')) {
			assert.throws(TypeError, function() { cm.char_to_string(-129); });
			assert.throws(TypeError, function() { cm.char_to_string(128); });
			assert.throws(TypeError, function() { cm.signed_char_to_string(-129); });
			assert.throws(TypeError, function() { cm.signed_char_to_string(128); });
			assert.throws(TypeError, function() { cm.unsigned_char_to_string(-1); });
			assert.throws(TypeError, function() { cm.unsigned_char_to_string(256); });
			assert.throws(TypeError, function() { cm.short_to_string(-32769); });
			assert.throws(TypeError, function() { cm.short_to_string(32768); });
			assert.throws(TypeError, function() { cm.unsigned_short_to_string(-1); });
			assert.throws(TypeError, function() { cm.unsigned_short_to_string(65536); });
			assert.throws(TypeError, function() { cm.int_to_string(-2147483649); });
			assert.throws(TypeError, function() { cm.int_to_string(2147483648); });
			assert.throws(TypeError, function() { cm.unsigned_int_to_string(-1); });
			assert.throws(TypeError, function() { cm.unsigned_int_to_string(4294967296); });
			assert.throws(TypeError, function() { cm.long_to_string(-2147483649); });
			assert.throws(TypeError, function() { cm.long_to_string(2147483648); });
			assert.throws(TypeError, function() { cm.unsigned_long_to_string(-1); });
			assert.throws(TypeError, function() { cm.unsigned_long_to_string(4294967296); });
			} else {
			// test that an out of range value doesn't throw without assertions.
			assert.equal("-129", cm.char_to_string(-129));
			}
			});

			test("unsigned values are correctly returned when stored in memory", function() {
			cm.store_unsigned_char(255);
			assert.equal(255, cm.load_unsigned_char());

			cm.store_unsigned_short(32768);
			assert.equal(32768, cm.load_unsigned_short());

			cm.store_unsigned_int(2147483648);
			assert.equal(2147483648, cm.load_unsigned_int());

			cm.store_unsigned_long(2147483648);
			assert.equal(2147483648, cm.load_unsigned_long());
			});

			if (cm.getCompilerSetting('ASSERTIONS')) {
			test("throws type error when attempting to coerce null to int", function() {
			var e = assert.throws(TypeError, function() {
			cm.int_to_string(null);
			});
			assert.equal('Cannot convert "null" to int', e.message);
			});
			} else {
			test("null is converted to 0 without assertions", function() {
			assert.equal('0', cm.int_to_string(null));
			});
			}

			test("access multiple class ctors", function() {
			var a = new cm.MultipleCtors(10);
			assert.equal(a.WhichCtorCalled(), 1);
			var b = new cm.MultipleCtors(20, 20);
			assert.equal(b.WhichCtorCalled(), 2);
			var c = new cm.MultipleCtors(30, 30, 30);
			assert.equal(c.WhichCtorCalled(), 3);
			a.delete();
			b.delete();
			c.delete();
			});

			test("access multiple smart ptr ctors", function() {
			var a = new cm.MultipleSmartCtors(10);
			assert.equal(a.WhichCtorCalled(), 1);
			var b = new cm.MultipleCtors(20, 20);
			assert.equal(b.WhichCtorCalled(), 2);
			a.delete();
			b.delete();
			});

			test("wrong number of constructor arguments throws", function() {
			assert.throws(cm.BindingError, function() { new cm.MultipleCtors(); });
			assert.throws(cm.BindingError, function() { new cm.MultipleCtors(1,2,3,4); });
			});

			test("overloading of free functions", function() {
			var a = cm.overloaded_function(10);
			assert.equal(a, 1);
			var b = cm.overloaded_function(20, 20);
			assert.equal(b, 2);
			});

			test("wrong number of arguments to an overloaded free function", function() {
			assert.throws(cm.BindingError, function() { cm.overloaded_function(); });
			assert.throws(cm.BindingError, function() { cm.overloaded_function(30, 30, 30); });
			});

			test("overloading of class member functions", function() {
			var foo = new cm.MultipleOverloads();
			assert.equal(foo.Func(10), 1);
			assert.equal(foo.WhichFuncCalled(), 1);
			assert.equal(foo.Func(20, 20), 2);
			assert.equal(foo.WhichFuncCalled(), 2);
			foo.delete();
			});

			test("wrong number of arguments to an overloaded class member function", function() {
			var foo = new cm.MultipleOverloads();
			assert.throws(cm.BindingError, function() { foo.Func(); });
			assert.throws(cm.BindingError, function() { foo.Func(30, 30, 30); });
			foo.delete();
			});

			test("wrong number of arguments to an overloaded class static function", function() {
			assert.throws(cm.BindingError, function() { cm.MultipleOverloads.StaticFunc(); });
			assert.throws(cm.BindingError, function() { cm.MultipleOverloads.StaticFunc(30, 30, 30); });
			});

			test("overloading of derived class member functions", function() {
			var foo = new cm.MultipleOverloadsDerived();

			// NOTE: In C++, default lookup rules will hide overloads from base class if derived class creates them.
			// In JS, we make the base class overloads implicitly available. In C++, they would need to be explicitly
			// invoked, like foo.MultipleOverloads::Func(10);
			assert.equal(foo.Func(10), 1);
			assert.equal(foo.WhichFuncCalled(), 1);
			assert.equal(foo.Func(20, 20), 2);
			assert.equal(foo.WhichFuncCalled(), 2);

			assert.equal(foo.Func(30, 30, 30), 3);
			assert.equal(foo.WhichFuncCalled(), 3);
			assert.equal(foo.Func(40, 40, 40, 40), 4);
			assert.equal(foo.WhichFuncCalled(), 4);
			foo.delete();
			});

			test("overloading of class static functions", function() {
			assert.equal(cm.MultipleOverloads.StaticFunc(10), 1);
			assert.equal(cm.MultipleOverloads.WhichStaticFuncCalled(), 1);
			assert.equal(cm.MultipleOverloads.StaticFunc(20, 20), 2);
			assert.equal(cm.MultipleOverloads.WhichStaticFuncCalled(), 2);
			});

			test("overloading of derived class static functions", function() {
			assert.equal(cm.MultipleOverloadsDerived.StaticFunc(30, 30, 30), 3);
			// TODO: Cannot access static member functions of a Base class via Derived.
			//            assert.equal(cm.MultipleOverloadsDerived.WhichStaticFuncCalled(), 3);
			assert.equal(cm.MultipleOverloads.WhichStaticFuncCalled(), 3);
			assert.equal(cm.MultipleOverloadsDerived.StaticFunc(40, 40, 40, 40), 4);
			// TODO: Cannot access static member functions of a Base class via Derived.
			//            assert.equal(cm.MultipleOverloadsDerived.WhichStaticFuncCalled(), 4);
			assert.equal(cm.MultipleOverloads.WhichStaticFuncCalled(), 4);
			});

			test("class member function named with a well-known symbol", function() {
			var instance = new cm.SymbolNameClass();
			assert.equal("Iterator", instance[Symbol.iterator]());
			assert.equal("Species", cm.SymbolNameClass[Symbol.species]());
			});

			test("no undefined entry in overload table when depending on already bound types", function() {
			var dummy_overloads = cm.MultipleOverloadsDependingOnDummy.prototype.dummy;
			// check if the overloadTable is correctly named
			// it can be minimized if using closure compiler
			if (dummy_overloads.hasOwnProperty('overloadTable')) {
			assert.false(dummy_overloads.overloadTable.hasOwnProperty('undefined'));
			}

			var dummy_static_overloads = cm.MultipleOverloadsDependingOnDummy.staticDummy;
			// check if the overloadTable is correctly named
			// it can be minimized if using closure compiler
			if (dummy_static_overloads.hasOwnProperty('overloadTable')) {
			assert.false(dummy_static_overloads.overloadTable.hasOwnProperty('undefined'));
			}

			// this part should fail anyway if there is no overloadTable
			var dependOnDummy = new cm.MultipleOverloadsDependingOnDummy();
			var dummy = dependOnDummy.dummy();
			dependOnDummy.dummy(dummy);
			dummy.delete();
			dependOnDummy.delete();

			// this part should fail anyway if there is no overloadTable
			var dummy = cm.MultipleOverloadsDependingOnDummy.staticDummy();
			cm.MultipleOverloadsDependingOnDummy.staticDummy(dummy);
			dummy.delete();
			});

			test("no undefined entry in overload table for free functions", function() {
			var dummy_free_func = cm.getDummy;
			console.log(dummy_free_func);

			if (dummy_free_func.hasOwnProperty('overloadTable')) {
			assert.false(dummy_free_func.overloadTable.hasOwnProperty('undefined'));
			}

			var dummy = cm.getDummy();
			cm.getDummy(dummy);
			});
		*/
	})

	When("vector", func() {
		It("std::vector returns as an native object", func() {
			vec, err := generated.Emval_test_return_vector(engine, ctx)

			size, err := vec.CallInstanceMethod(ctx, vec, "size")
			Expect(err).To(BeNil())
			Expect(size).To(Equal(uint32(3)))

			get0, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(0))
			Expect(err).To(BeNil())
			Expect(get0).To(Equal(int32(10)))

			get1, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(1))
			Expect(err).To(BeNil())
			Expect(get1).To(Equal(int32(20)))

			get2, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(2))
			Expect(err).To(BeNil())
			Expect(get2).To(Equal(int32(30)))

			err = vec.DeleteInstance(ctx, vec)
			Expect(err).To(BeNil())
		})

		It("out of bounds std::vector access returns undefined", func() {
			vec, err := generated.Emval_test_return_vector(engine, ctx)

			get4, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(4))
			Expect(err).To(BeNil())
			Expect(get4).To(Equal(types.Undefined))

			err = vec.DeleteInstance(ctx, vec)
			Expect(err).To(BeNil())
		})

		It("std::vector<std::shared_ptr<>> can be passed back", func() {
			vec, err := generated.Emval_test_return_shared_ptr_vector(engine, ctx)
			Expect(err).To(BeNil())

			size, err := vec.CallInstanceMethod(ctx, vec, "size")
			Expect(err).To(BeNil())
			Expect(size).To(Equal(uint32(2)))

			str0, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(0))
			Expect(err).To(BeNil())

			str1, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(1))
			Expect(err).To(BeNil())

			str0Str, err := str0.(embind.IClassBase).CallInstanceMethod(ctx, str0, "get")
			Expect(err).To(BeNil())
			Expect(str0Str).To(Equal("string #1"))

			str1Str, err := str1.(embind.IClassBase).CallInstanceMethod(ctx, str1, "get")
			Expect(err).To(BeNil())
			Expect(str1Str).To(Equal("string #2"))

			err = str0.(embind.IClassBase).DeleteInstance(ctx, str0.(embind.IClassBase))
			Expect(err).To(BeNil())

			err = str1.(embind.IClassBase).DeleteInstance(ctx, str1.(embind.IClassBase))
			Expect(err).To(BeNil())

			err = vec.DeleteInstance(ctx, vec)
			Expect(err).To(BeNil())
		})

		It("objects can be pushed back", func() {
			vectorHolder, err := generated.NewClassVectorHolder(engine, ctx)
			Expect(err).To(BeNil())

			vec, err := vectorHolder.Get(ctx)
			Expect(err).To(BeNil())

			size, err := vec.CallInstanceMethod(ctx, vec, "size")
			Expect(size).To(Equal(uint32(2)))

			str, err := generated.NewClassStringHolder(engine, ctx, "abc")
			Expect(err).To(BeNil())

			_, err = vec.CallInstanceMethod(ctx, vec, "push_back", str)
			Expect(err).To(BeNil())

			err = str.Delete(ctx)
			Expect(err).To(BeNil())

			size, err = vec.CallInstanceMethod(ctx, vec, "size")
			Expect(err).To(BeNil())
			Expect(size).To(Equal(uint32(3)))

			getStr, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(2))
			Expect(err).To(BeNil())

			getStrGet, err := getStr.(embind.IClassBase).CallInstanceMethod(ctx, getStr, "get")
			Expect(err).To(BeNil())
			Expect(getStrGet).To(Equal("abc"))

			err = getStr.(embind.IClassBase).DeleteInstance(ctx, getStr.(embind.IClassBase))
			Expect(err).To(BeNil())

			err = vec.DeleteInstance(ctx, vec)
			Expect(err).To(BeNil())

			err = vectorHolder.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("can get elements with array operator", func() {
			vec, err := generated.Emval_test_return_vector(engine, ctx)
			Expect(err).To(BeNil())

			get, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(0))
			Expect(err).To(BeNil())
			Expect(get).To(Equal(int32(10)))

			err = vec.DeleteInstance(ctx, vec)
			Expect(err).To(BeNil())
		})

		It("can set elements with array operator", func() {
			vec, err := generated.Emval_test_return_vector(engine, ctx)
			Expect(err).To(BeNil())

			get0, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(0))
			Expect(err).To(BeNil())
			Expect(get0).To(Equal(int32(10)))

			_, err = vec.CallInstanceMethod(ctx, vec, "set", uint32(2), int32(60))
			Expect(err).To(BeNil())

			get2, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(2))
			Expect(err).To(BeNil())
			Expect(get2).To(Equal(int32(60)))

			err = vec.DeleteInstance(ctx, vec)
			Expect(err).To(BeNil())
		})

		It("can set and get objects", func() {
			vec, err := generated.Emval_test_return_shared_ptr_vector(engine, ctx)
			Expect(err).To(BeNil())

			str, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(0))
			Expect(err).To(BeNil())

			strGetStr, err := str.(embind.IClassBase).CallInstanceMethod(ctx, str.(embind.IClassBase), "get")
			Expect(err).To(BeNil())
			Expect(strGetStr).To(Equal("string #1"))

			err = str.(embind.IClassBase).DeleteInstance(ctx, str.(embind.IClassBase))
			Expect(err).To(BeNil())

			err = vec.DeleteInstance(ctx, vec)
			Expect(err).To(BeNil())
		})

		It("resize appends the given value", func() {
			vec, err := generated.Emval_test_return_vector(engine, ctx)
			Expect(err).To(BeNil())

			_, err = vec.CallInstanceMethod(ctx, vec, "resize", uint32(5), int32(42))
			Expect(err).To(BeNil())

			size, err := vec.CallInstanceMethod(ctx, vec, "size")
			Expect(err).To(BeNil())
			Expect(size).To(Equal(uint32(5)))

			get0, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(0))
			Expect(err).To(BeNil())
			Expect(get0).To(Equal(int32(10)))

			get1, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(1))
			Expect(err).To(BeNil())
			Expect(get1).To(Equal(int32(20)))

			get2, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(2))
			Expect(err).To(BeNil())
			Expect(get2).To(Equal(int32(30)))

			get3, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(3))
			Expect(err).To(BeNil())
			Expect(get3).To(Equal(int32(42)))

			get4, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(4))
			Expect(err).To(BeNil())
			Expect(get4).To(Equal(int32(42)))

			err = vec.DeleteInstance(ctx, vec)
			Expect(err).To(BeNil())
		})

		It("resize preserves content when shrinking", func() {
			vec, err := generated.Emval_test_return_vector(engine, ctx)
			Expect(err).To(BeNil())

			_, err = vec.CallInstanceMethod(ctx, vec, "resize", uint32(2), int32(42))
			Expect(err).To(BeNil())

			size, err := vec.CallInstanceMethod(ctx, vec, "size")
			Expect(err).To(BeNil())
			Expect(size).To(Equal(uint32(2)))

			get0, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(0))
			Expect(err).To(BeNil())
			Expect(get0).To(Equal(int32(10)))

			get1, err := vec.CallInstanceMethod(ctx, vec, "get", uint32(1))
			Expect(err).To(BeNil())
			Expect(get1).To(Equal(int32(20)))

			err = vec.DeleteInstance(ctx, vec)
			Expect(err).To(BeNil())
		})
	})

	When("map", func() {
		It("std::map returns as native object", func() {
			newMap, err := generated.Embind_test_get_string_int_map(engine, ctx)
			Expect(err).To(BeNil())

			size, err := newMap.CallInstanceMethod(ctx, newMap, "size")
			Expect(err).To(BeNil())
			Expect(size).To(Equal(uint32(2)))

			one, err := newMap.CallInstanceMethod(ctx, newMap, "get", "one")
			Expect(err).To(BeNil())
			Expect(one).To(Equal(int32(1)))

			two, err := newMap.CallInstanceMethod(ctx, newMap, "get", "two")
			Expect(err).To(BeNil())
			Expect(two).To(Equal(int32(2)))

			err = newMap.DeleteInstance(ctx, newMap)
			Expect(err).To(BeNil())
		})

		It("std::map can get keys", func() {
			newMap, err := generated.Embind_test_get_string_int_map(engine, ctx)
			Expect(err).To(BeNil())

			size, err := newMap.CallInstanceMethod(ctx, newMap, "size")
			Expect(err).To(BeNil())

			keys, err := newMap.CallInstanceMethod(ctx, newMap, "keys")
			Expect(err).To(BeNil())

			keysClass := keys.(embind.IClassBase)

			keysSize, err := keysClass.CallInstanceMethod(ctx, keysClass, "size")
			Expect(err).To(BeNil())

			Expect(keysSize).To(Equal(size.(uint32)))

			one, err := keysClass.CallInstanceMethod(ctx, keysClass, "get", uint32(0))
			Expect(err).To(BeNil())
			Expect(one).To(Equal("one"))

			two, err := keysClass.CallInstanceMethod(ctx, keysClass, "get", uint32(1))
			Expect(err).To(BeNil())
			Expect(two).To(Equal("two"))

			err = keysClass.DeleteInstance(ctx, keysClass)

			err = newMap.DeleteInstance(ctx, newMap)
			Expect(err).To(BeNil())
		})

		It("std::map can set keys and values", func() {
			newMap, err := generated.Embind_test_get_string_int_map(engine, ctx)

			size, err := newMap.CallInstanceMethod(ctx, newMap, "size")
			Expect(err).To(BeNil())
			Expect(size).To(Equal(uint32(2)))

			_, err = newMap.CallInstanceMethod(ctx, newMap, "set", "three", int32(3))
			Expect(err).To(BeNil())

			size, err = newMap.CallInstanceMethod(ctx, newMap, "size")
			Expect(err).To(BeNil())
			Expect(size).To(Equal(uint32(3)))

			three, err := newMap.CallInstanceMethod(ctx, newMap, "get", "three")
			Expect(err).To(BeNil())
			Expect(three).To(Equal(int32(3)))

			_, err = newMap.CallInstanceMethod(ctx, newMap, "set", "three", int32(4))
			Expect(err).To(BeNil())

			size, err = newMap.CallInstanceMethod(ctx, newMap, "size")
			Expect(err).To(BeNil())
			Expect(size).To(Equal(uint32(3)))

			three, err = newMap.CallInstanceMethod(ctx, newMap, "get", "three")
			Expect(err).To(BeNil())
			Expect(three).To(Equal(int32(4)))

			err = newMap.DeleteInstance(ctx, newMap)
			Expect(err).To(BeNil())
		})
	})

	When("functors", func() {
		It("can get and call function ptrs", func() {
			ptr, err := generated.Emval_test_get_function_ptr(engine, ctx)
			Expect(err).To(BeNil())

			opcall, err := ptr.CallInstanceMethod(ctx, ptr, "opcall", "foobar")
			Expect(err).To(BeNil())
			Expect(opcall).To(Equal("foobar"))

			err = ptr.DeleteInstance(ctx, ptr)
			Expect(err).To(BeNil())
		})

		It("can pass functor to C++", func() {
			ptr, err := generated.Emval_test_get_function_ptr(engine, ctx)
			Expect(err).To(BeNil())

			takeAndCallResult, err := generated.Emval_test_take_and_call_functor(engine, ctx, ptr)
			Expect(err).To(BeNil())
			Expect(takeAndCallResult).To(Equal("asdf"))

			err = ptr.DeleteInstance(ctx, ptr)
			Expect(err).To(BeNil())
		})

		It("can clone handles", func() {
			a, err := generated.Emval_test_get_function_ptr(engine, ctx)
			Expect(err).To(BeNil())

			b, err := a.CloneInstance(ctx, a)
			Expect(err).To(BeNil())

			err = a.DeleteInstance(ctx, a)
			Expect(err).To(BeNil())

			err = a.DeleteInstance(ctx, a)
			Expect(err).To(Not(BeNil()))
			Expect(err.Error()).To(ContainSubstring("class handle already deleted"))

			err = b.DeleteInstance(ctx, b)
			Expect(err).To(BeNil())
		})
	})

	When("classes", func() {
		It("class instance", func() {
			a := map[string]any{"foo": "bar"}

			countEmvalHandles := engine.CountEmvalHandles()
			Expect(countEmvalHandles).To(Equal(0))

			c, err := generated.NewClassValHolder(engine, ctx, a)
			Expect(err).To(BeNil())

			countEmvalHandles = engine.CountEmvalHandles()
			Expect(countEmvalHandles).To(Equal(1))

			getVal, err := c.GetVal(ctx)
			Expect(err).To(BeNil())
			Expect(getVal).To(HaveKeyWithValue("foo", "bar"))

			countEmvalHandles = engine.CountEmvalHandles()
			Expect(countEmvalHandles).To(Equal(1))

			err = c.SetVal(ctx, "1234")
			Expect(err).To(BeNil())

			getVal, err = c.GetVal(ctx)
			Expect(err).To(BeNil())
			Expect(getVal).To(Equal("1234"))

			err = c.Delete(ctx)
			Expect(err).To(BeNil())

			countEmvalHandles = engine.CountEmvalHandles()
			Expect(countEmvalHandles).To(Equal(0))
		})

		It("class properties can be methods", func() {
			a := map[string]any{}
			b := map[string]any{"foo": "foo"}
			c, err := generated.NewClassValHolder(engine, ctx, a)
			Expect(err).To(BeNil())

			val, err := c.GetPropertyVal(ctx)
			Expect(err).To(BeNil())
			Expect(val).To(Equal(a))

			err = c.SetPropertyVal(ctx, b)
			Expect(err).To(BeNil())

			val, err = c.GetPropertyVal(ctx)
			Expect(err).To(BeNil())
			Expect(val).To(Equal(b))

			err = c.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("class properties can be std::function objects", func() {
			a := map[string]any{}
			b := map[string]any{"foo": "foo"}
			c, err := generated.NewClassValHolder(engine, ctx, a)
			Expect(err).To(BeNil())

			val, err := c.GetPropertyFunction_val(ctx)
			Expect(err).To(BeNil())
			Expect(val).To(Equal(a))

			err = c.SetPropertyFunction_val(ctx, b)
			Expect(err).To(BeNil())

			val, err = c.GetPropertyFunction_val(ctx)
			Expect(err).To(BeNil())
			Expect(val).To(Equal(b))

			err = c.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("class properties can be read-only std::function objects", func() {
			a := map[string]any{}

			h, err := generated.NewClassValHolder(engine, ctx, a)
			Expect(err).To(BeNil())

			funcVal, err := h.GetPropertyReadonly_function_val(ctx)
			Expect(err).To(BeNil())
			Expect(funcVal).To(Equal(a))

			err = h.SetProperty(ctx, "readonly_function_val", 10)
			Expect(err).To(Not(BeNil()))
			Expect(err.Error()).To(ContainSubstring("ValHolder.readonly_function_val is a read-only property"))

			err = h.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("class properties can be function objects (functor)", func() {
			a := map[string]any{}
			b := map[string]any{"foo": "foo"}
			c, err := generated.NewClassValHolder(engine, ctx, a)
			Expect(err).To(BeNil())

			functor_val, err := c.GetPropertyFunctor_val(ctx)
			Expect(err).To(BeNil())
			Expect(functor_val).To(Equal(a))

			err = c.SetPropertyFunction_val(ctx, b)
			Expect(err).To(BeNil())

			functor_val, err = c.GetPropertyFunctor_val(ctx)
			Expect(err).To(BeNil())
			Expect(functor_val).To(Equal(b))

			err = c.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("class properties can be read-only function objects (functor)", func() {
			a := map[string]any{}
			h, err := generated.NewClassValHolder(engine, ctx, a)
			Expect(err).To(BeNil())

			readonly_functor_val, err := h.GetPropertyReadonly_functor_val(ctx)
			Expect(err).To(BeNil())
			Expect(readonly_functor_val).To(Equal(a))

			err = h.SetProperty(ctx, "readonly_functor_val", 10)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("ValHolder.readonly_functor_val is a read-only property"))

			err = h.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("class properties can be read-only", func() {
			a := map[string]any{}
			h, err := generated.NewClassValHolder(engine, ctx, a)
			Expect(err).To(BeNil())

			val_readonly, err := h.GetPropertyVal_readonly(ctx)
			Expect(err).To(BeNil())
			Expect(val_readonly).To(Equal(a))

			err = h.SetProperty(ctx, "val_readonly", 10)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("ValHolder.val_readonly is a read-only property"))

			err = h.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("read-only member field", func() {
			a, err := generated.NewClassHasReadOnlyProperty(engine, ctx, 10)

			i, err := a.GetPropertyI(ctx)
			Expect(i).To(Equal(int32(10)))

			err = a.SetProperty(ctx, "i", 20)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("HasReadOnlyProperty.i is a read-only property"))

			err = a.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("class instance $$ property is non-enumerable", func() {
			c, err := generated.NewClassValHolder(engine, ctx, types.Undefined)
			Expect(err).To(BeNil())

			//assert.deepEqual([], Object.keys(c));

			d, err := c.Clone(ctx)
			Expect(err).To(BeNil())

			err = c.Delete(ctx)
			Expect(err).To(BeNil())

			//assert.deepEqual([], Object.keys(d));

			err = d.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("class methods", func() {
			someClassMethod, err := generated.ClassValHolderStaticSome_class_method(engine, ctx, 10)
			Expect(err).To(BeNil())
			Expect(someClassMethod).To(Equal(int32(10)))

			b, err := generated.ClassValHolderStaticMakeValHolder(engine, ctx, "foo")
			Expect(err).To(BeNil())

			getVal, err := b.CallInstanceMethod(ctx, b, "getVal")
			Expect(err).To(BeNil())
			Expect(getVal).To(Equal("foo"))

			err = b.DeleteInstance(ctx, b)
			Expect(err).To(BeNil())
		})

		It("function objects as class constructors", func() {
			a, err := generated.NewClassConstructFromStdFunction(engine, ctx, "foo", 10)
			Expect(err).To(BeNil())

			getVal, err := a.GetVal(ctx)
			Expect(err).To(BeNil())
			Expect(getVal).To(Equal("foo"))

			getA, err := a.GetA(ctx)
			Expect(err).To(BeNil())
			Expect(getA).To(Equal(int32(10)))

			b, err := generated.NewClassConstructFromFunctionObject(engine, ctx, "bar", 12)
			Expect(err).To(BeNil())

			getVal, err = b.GetVal(ctx)
			Expect(err).To(BeNil())
			Expect(getVal).To(Equal("bar"))

			getA, err = b.GetA(ctx)
			Expect(err).To(BeNil())
			Expect(getA).To(Equal(int32(12)))

			err = a.Delete(ctx)
			Expect(err).To(BeNil())

			err = b.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("function objects as class methods", func() {
			b, err := generated.ClassValHolderStaticMakeValHolder(engine, ctx, "foo")
			Expect(err).To(BeNil())

			// get & set via std::function
			getValFunction, err := b.CallInstanceMethod(ctx, b, "getValFunction")
			Expect(err).To(BeNil())
			Expect(getValFunction).To(Equal("foo"))

			_, err = b.CallInstanceMethod(ctx, b, "setValFunction", "bar")
			Expect(err).To(BeNil())

			// get & set via 'callable'
			getValFunctor, err := b.CallInstanceMethod(ctx, b, "getValFunctor")
			Expect(err).To(BeNil())
			Expect(getValFunctor).To(Equal("bar"))

			_, err = b.CallInstanceMethod(ctx, b, "setValFunctor", "baz")
			Expect(err).To(BeNil())

			getValFunction, err = b.CallInstanceMethod(ctx, b, "getValFunction")
			Expect(err).To(BeNil())
			Expect(getValFunction).To(Equal("baz"))

			err = b.DeleteInstance(ctx, b)
			Expect(err).To(BeNil())
		})

		It("can't call methods on deleted class instances", func() {
			c, err := generated.NewClassValHolder(engine, ctx, types.Undefined)
			Expect(err).To(BeNil())

			err = c.Delete(ctx)
			Expect(err).To(BeNil())

			_, err = c.GetVal(ctx)
			Expect(err).To(Not(BeNil()))
			Expect(err.Error()).To(ContainSubstring("cannot pass deleted object as a pointer of type ValHolder const*"))

			err = c.Delete(ctx)
			Expect(err).To(Not(BeNil()))
			Expect(err.Error()).To(ContainSubstring("class handle already deleted"))
		})

		It("can return class instances by value", func() {
			c, err := generated.Emval_test_return_ValHolder(engine, ctx)
			Expect(err).To(BeNil())

			getVal, err := c.CallInstanceMethod(ctx, c, "getVal")
			Expect(err).To(BeNil())
			Expect(getVal).To(Equal(map[string]any{}))

			err = c.DeleteInstance(ctx, c)
			Expect(err).To(BeNil())
		})

		It("can pass class instances to functions by reference", func() {
			a := map[string]any{"a": 1}
			c, err := generated.NewClassValHolder(engine, ctx, a)
			err = generated.Emval_test_set_ValHolder_to_empty_object(engine, ctx, c)
			getVal, err := c.CallInstanceMethod(ctx, c, "getVal")
			Expect(err).To(BeNil())
			Expect(getVal).To(Equal(map[string]any{}))

			err = c.DeleteInstance(ctx, c)
			Expect(err).To(BeNil())
		})

		It("can pass smart pointer by reference", func() {
			base, err := generated.Embind_test_return_smart_base_ptr(engine, ctx)
			Expect(err).To(BeNil())

			name, err := generated.Embind_test_get_class_name_via_reference_to_smart_base_ptr(engine, ctx, base)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("Base"))

			err = base.DeleteInstance(ctx, base)
			Expect(err).To(BeNil())
		})

		It("can pass smart pointer by value", func() {
			base, err := generated.Embind_test_return_smart_base_ptr(engine, ctx)
			Expect(err).To(BeNil())

			name, err := generated.Embind_test_get_class_name_via_smart_base_ptr(engine, ctx, base)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("Base"))

			err = base.DeleteInstance(ctx, base)
			Expect(err).To(BeNil())
		})

		// todo: fix this (comment from Emscripten
		// This test does not work because we make no provision for argument values
		// having been changed after returning from a C++ routine invocation. In
		// this specific case, the original pointee of the smart pointer was
		// freed and replaced by a new one, but the ptr in our local handle
		// was never updated after returning from the call.
		It("can modify smart pointers passed by reference", func() {
			//            var base = cm.embind_test_return_smart_base_ptr();
			//            cm.embind_modify_smart_pointer_passed_by_reference(base);
			//            assert.equal("Changed", base.getClassName());
			//            base.delete();
		})

		It("can not modify smart pointers passed by value", func() {
			base, err := generated.Embind_test_return_smart_base_ptr(engine, ctx)
			Expect(err).To(BeNil())

			err = generated.Embind_attempt_to_modify_smart_pointer_when_passed_by_value(engine, ctx, base)
			Expect(err).To(BeNil())

			className, err := base.CallInstanceMethod(ctx, base, "getClassName")
			Expect(err).To(BeNil())
			Expect(className).To(Equal("Base"))

			err = base.DeleteInstance(ctx, base)
			Expect(err).To(BeNil())
		})

		It("const return value", func() {
			c, err := generated.NewClassValHolder(engine, ctx, "foo")
			Expect(err).To(BeNil())

			constRef, err := c.GetConstVal(ctx)
			Expect(err).To(BeNil())
			Expect(constRef).To(Equal("foo"))

			err = c.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("return object by const ref", func() {
			c, err := generated.NewClassValHolder(engine, ctx, "foo")
			Expect(err).To(BeNil())

			constRef, err := c.GetValConstRef(ctx)
			Expect(err).To(BeNil())
			Expect(constRef).To(Equal("foo"))

			err = c.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("instanceof", func() {
			c, err := generated.NewClassValHolder(engine, ctx, "foo")
			Expect(err).To(BeNil())

			err = c.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("can access struct fields", func() {
			c, err := generated.NewClassCustomStruct(engine, ctx)
			Expect(err).To(BeNil())

			field, err := c.GetPropertyField(ctx)
			Expect(err).To(BeNil())
			Expect(field).To(Equal(int32(10)))

			field, err = c.GetField(ctx)
			Expect(err).To(BeNil())
			Expect(field).To(Equal(int32(10)))

			err = c.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("can set struct fields", func() {
			c, err := generated.NewClassCustomStruct(engine, ctx)
			Expect(err).To(BeNil())

			err = c.SetPropertyField(ctx, 15)
			Expect(err).To(BeNil())

			field, err := c.GetPropertyField(ctx)
			Expect(err).To(BeNil())
			Expect(field).To(Equal(int32(15)))

			err = c.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("can return tuples by value", func() {
			c, err := generated.Emval_test_return_TupleVector(engine, ctx)
			Expect(err).To(BeNil())
			Expect(c).To(Equal([]any{float32(1), float32(2), float32(3), float32(4)}))
		})

		It("tuples can contain tuples", func() {
			c, err := generated.Emval_test_return_TupleVectorTuple(engine, ctx)
			Expect(err).To(BeNil())
			Expect(c).To(Equal([]any{[]any{float32(1), float32(2), float32(3), float32(4)}}))
		})

		It("can pass tuples by value", func() {
			c, err := generated.Emval_test_take_and_return_TupleVector(engine, ctx, []any{float32(4), float32(5), float32(6), float32(7)})
			Expect(err).To(BeNil())
			Expect(c).To(Equal([]any{float32(4), float32(5), float32(6), float32(7)}))
		})

		It("can return structs by value", func() {
			c, err := generated.Emval_test_return_StructVector(engine, ctx)
			Expect(err).To(BeNil())
			Expect(c).To(Equal(map[string]any{"x": float32(1), "y": float32(2), "z": float32(3), "w": float32(4)}))
		})

		It("can pass structs by value", func() {
			c, err := generated.Emval_test_take_and_return_StructVector(engine, ctx, map[string]any{"x": float32(4), "y": float32(5), "z": float32(6), "w": float32(7)})
			Expect(err).To(BeNil())
			Expect(c).To(Equal(map[string]any{"x": float32(4), "y": float32(5), "z": float32(6), "w": float32(7)}))
		})

		It("can pass and return tuples in structs", func() {
			d, err := generated.Emval_test_take_and_return_TupleInStruct(engine, ctx, map[string]any{"field": []any{float32(1), float32(2), float32(3), float32(4)}})
			Expect(err).To(BeNil())
			Expect(d).To(Equal(map[string]any{"field": []any{float32(1), float32(2), float32(3), float32(4)}}))
		})

		It("can pass and return arrays in structs", func() {
			d, err := generated.Emval_test_take_and_return_ArrayInStruct(engine, ctx, map[string]any{
				"field1": []any{int32(1), int32(2)},
				"field2": []any{
					map[string]any{"x": int32(1), "y": int32(2)},
					map[string]any{"x": int32(3), "y": int32(4)},
				},
			})
			Expect(err).To(BeNil())
			Expect(d).To(Equal(map[string]any{
				"field1": []any{int32(1), int32(2)},
				"field2": []any{
					map[string]any{"x": int32(1), "y": int32(2)},
					map[string]any{"x": int32(3), "y": int32(4)},
				},
			}))
		})

		It("can clone handles", func() {
			a, err := generated.NewClassValHolder(engine, ctx, struct{}{})
			countHandles := engine.CountEmvalHandles()
			Expect(err).To(BeNil())
			Expect(countHandles).To(Equal(1))

			b, err := a.Clone(ctx)
			Expect(err).To(BeNil())

			err = a.DeleteInstance(ctx, a)
			Expect(err).To(BeNil())

			countHandles = engine.CountEmvalHandles()
			Expect(countHandles).To(Equal(1))

			err = a.DeleteInstance(ctx, a)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("class handle already deleted"))

			err = b.DeleteInstance(ctx, b)
			Expect(err).To(BeNil())

			countHandles = engine.CountEmvalHandles()
			Expect(countHandles).To(Equal(0))
		})

		It("A shared pointer set/get point to the same underlying pointer", func() {
			a, err := generated.NewClassSharedPtrHolder(engine, ctx)
			Expect(err).To(BeNil())

			b, err := a.Get(ctx)
			Expect(err).To(BeNil())

			err = a.Set(ctx, b)
			Expect(err).To(BeNil())

			c, err := a.Get(ctx)
			Expect(err).To(BeNil())

			isAliasOf, err := b.IsAliasOfInstance(ctx, b, c)
			Expect(err).To(BeNil())
			Expect(isAliasOf).To(BeTrue())

			err = b.DeleteInstance(ctx, b)
			Expect(err).To(BeNil())

			err = c.DeleteInstance(ctx, c)
			Expect(err).To(BeNil())

			err = a.DeleteInstance(ctx, a)
			Expect(err).To(BeNil())
		})

		It("can return shared ptrs from instance methods", func() {
			a, err := generated.NewClassSharedPtrHolder(engine, ctx)
			Expect(err).To(BeNil())

			b, err := a.Get(ctx)
			Expect(err).To(BeNil())

			get, err := b.CallInstanceMethod(ctx, b, "get")
			Expect(get).To(Equal("a string"))

			err = b.DeleteInstance(ctx, b)
			Expect(err).To(BeNil())

			err = a.DeleteInstance(ctx, a)
			Expect(err).To(BeNil())
		})

		It("smart ptrs clone correctly", func() {
			countHandles := engine.CountEmvalHandles()
			Expect(countHandles).To(Equal(0))

			a, err := generated.Emval_test_return_shared_ptr(engine, ctx)
			Expect(err).To(BeNil())

			b, err := a.CloneInstance(ctx, a)
			Expect(err).To(BeNil())

			err = a.DeleteInstance(ctx, a)
			Expect(err).To(BeNil())

			countHandles = engine.CountEmvalHandles()
			Expect(countHandles).To(Equal(1))

			err = a.DeleteInstance(ctx, a)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("as"))

			err = b.DeleteInstance(ctx, b)
			Expect(err).To(BeNil())

			countHandles = engine.CountEmvalHandles()
			Expect(countHandles).To(Equal(0))
		})

		It("can't clone if already deleted", func() {
			a, err := generated.NewClassValHolder(engine, ctx, struct{}{})
			Expect(err).To(BeNil())

			err = a.Delete(ctx)
			Expect(err).To(BeNil())

			_, err = a.Clone(ctx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("class handle already deleted"))
		})

		It("virtual calls work correctly", func() {
			derived, err := generated.Embind_test_return_raw_polymorphic_derived_ptr_as_base(engine, ctx)
			Expect(err).To(BeNil())

			virtualClassName, err := derived.CallInstanceMethod(ctx, derived, "virtualGetClassName")
			Expect(err).To(BeNil())
			Expect(virtualClassName).To(Equal("PolyDerived"))

			err = derived.DeleteInstance(ctx, derived)
			Expect(err).To(BeNil())
		})

		It("virtual calls work correctly on smart ptrs", func() {
			derived, err := generated.Embind_test_return_smart_polymorphic_derived_ptr_as_base(engine, ctx)
			Expect(err).To(BeNil())

			virtualClassName, err := derived.CallInstanceMethod(ctx, derived, "virtualGetClassName")
			Expect(err).To(BeNil())
			Expect(virtualClassName).To(Equal("PolyDerived"))

			err = derived.DeleteInstance(ctx, derived)
			Expect(err).To(BeNil())
		})

		It("Empty smart ptr is null", func() {
			a, err := generated.Emval_test_return_empty_shared_ptr(engine, ctx)
			Expect(err).To(BeNil())
			Expect(a).To(BeNil())
		})

		It("string cannot be given as smart pointer argument", func() {
			_, err := engine.CallPublicSymbol(ctx, "emval_test_is_shared_ptr_null", "hello world")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("check whether you constructed it properly through embind, the given value is a string"))
		})

		It("number cannot be given as smart pointer argument", func() {
			_, err := engine.CallPublicSymbol(ctx, "emval_test_is_shared_ptr_null", 105)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("check whether you constructed it properly through embind, the given value is a int"))
		})

		It("raw pointer cannot be given as smart pointer argument", func() {
			p, err := generated.NewClassValHolder(engine, ctx, struct{}{})
			Expect(err).To(BeNil())

			_, err = generated.Emval_test_is_shared_ptr_null(engine, ctx, p)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("passing raw pointer to smart pointer is illegal"))

			err = p.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("null is passed as empty smart pointer", func() {
			isNull, err := generated.Emval_test_is_shared_ptr_null(engine, ctx, nil)
			Expect(err).To(BeNil())
			Expect(isNull).To(BeTrue())
		})

		It("Deleting already deleted smart ptrs fails", func() {
			a, err := generated.Emval_test_return_shared_ptr(engine, ctx)
			Expect(err).To(BeNil())

			err = a.DeleteInstance(ctx, a)
			Expect(err).To(BeNil())

			err = a.DeleteInstance(ctx, a)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("class handle already deleted"))
		})

		It("returned unique_ptr does not call destructor", func() {
			logged := ""

			c, err := generated.Emval_test_return_unique_ptr_lifetime(engine, ctx, func(s string) { logged += s })
			Expect(err).To(BeNil())
			Expect(logged).To(Equal("(constructor)"))

			err = c.DeleteInstance(ctx, c)
			Expect(err).To(BeNil())
		})

		It("returned unique_ptr calls destructor on delete", func() {
			logged := ""

			c, err := generated.Emval_test_return_unique_ptr_lifetime(engine, ctx, func(s string) { logged += s })
			Expect(err).To(BeNil())

			logged = ""

			err = c.DeleteInstance(ctx, c)
			Expect(err).To(BeNil())

			Expect(logged).To(Equal("(destructor)"))
		})

		It("StringHolder", func() {
			a, err := generated.NewClassStringHolder(engine, ctx, "foobar")
			Expect(err).To(BeNil())

			str, err := a.Get(ctx)
			Expect(err).To(BeNil())
			Expect(str).To(Equal("foobar"))

			err = a.Set(ctx, "barfoo")

			str, err = a.Get(ctx)
			Expect(err).To(BeNil())
			Expect(str).To(Equal("barfoo"))

			constRef, err := a.Get_const_ref(ctx)
			Expect(err).To(BeNil())
			Expect(constRef).To(Equal("barfoo"))

			err = a.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("can call methods on unique ptr", func() {
			result, err := generated.Emval_test_return_unique_ptr(engine, ctx)
			Expect(err).To(BeNil())

			_, err = result.CallInstanceMethod(ctx, result, "setVal", "1234")
			Expect(err).To(BeNil())

			getVal, err := result.CallInstanceMethod(ctx, result, "getVal")
			Expect(err).To(BeNil())
			Expect(getVal).To(Equal("1234"))

			err = result.DeleteInstance(ctx, result)
			Expect(err).To(BeNil())
		})

		It("can call methods on shared ptr", func() {
			result, err := generated.Emval_test_return_shared_ptr(engine, ctx)
			Expect(err).To(BeNil())

			_, err = result.CallInstanceMethod(ctx, result, "setVal", "1234")
			Expect(err).To(BeNil())

			getVal, err := result.CallInstanceMethod(ctx, result, "getVal")
			Expect(err).To(BeNil())
			Expect(getVal).To(Equal("1234"))

			err = result.DeleteInstance(ctx, result)
			Expect(err).To(BeNil())
		})

		It("non-member methods", func() {
			a := map[string]any{"foo": "bar"}
			c, err := generated.NewClassValHolder(engine, ctx, a)
			Expect(err).To(BeNil())

			err = c.SetEmpty(ctx)
			Expect(err).To(BeNil())

			valNonMember, err := c.GetValNonMember(ctx)
			Expect(err).To(BeNil())
			Expect(valNonMember).To(Equal(map[string]any{}))

			err = c.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("instantiating class without constructor gives error", func() {
			_, err := generated.AbstractClass(engine, ctx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("AbstractClass has no accessible constructor"))
		})

		It("can construct class with external constructor", func() {
			e, err := generated.NewClassHasExternalConstructor(engine, ctx, "foo")
			Expect(err).To(BeNil())

			str, err := e.GetString(ctx)
			Expect(err).To(BeNil())
			Expect(str).To(Equal("foo"))

			err = e.Delete(ctx)
			Expect(err).To(BeNil())
		})
	})

	When("const", func() {
		It("calling non-const method with const handle is error", func() {
			vh, err := generated.ClassValHolderStaticMakeConst(engine, ctx, struct{}{})
			Expect(err).To(BeNil())

			_, err = vh.CallInstanceMethod(ctx, vh, "setVal", struct{}{})
			Expect(err).To(Not(BeNil()))
			Expect(err.Error()).To(ContainSubstring("cannot convert argument of type ValHolder const* to parameter type ValHolder*"))

			err = vh.DeleteInstance(ctx, vh)
			Expect(err).To(BeNil())
		})

		It("passing const pointer to non-const pointer is error", func() {
			vh, err := generated.ClassValHolderStaticMakeConst(engine, ctx, struct{}{})

			err = generated.ClassValHolderStaticSet_via_raw_pointer(engine, ctx, vh, struct{}{})
			Expect(err).To(Not(BeNil()))

			Expect(err.Error()).To(ContainSubstring("cannot convert argument of type ValHolder const* to parameter type ValHolder*"))

			err = vh.DeleteInstance(ctx, vh)
			Expect(err).To(BeNil())
		})
	})

	When("smart pointers", func() {
		It("constructor can return smart pointer", func() {
			e, err := generated.NewClassHeldBySmartPtr(engine, ctx, 10, "foo")
			Expect(err).To(BeNil())

			i, err := e.GetPropertyI(ctx)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(int32(10)))

			s, err := e.GetPropertyS(ctx)
			Expect(err).To(BeNil())
			Expect(s).To(Equal("foo"))

			f, err := generated.TakesHeldBySmartPtr(engine, ctx, e)
			Expect(err).To(BeNil())

			err = f.DeleteInstance(ctx, f)
			Expect(err).To(BeNil())

			err = e.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("cannot pass incorrect smart pointer type", func() {
			e, err := generated.Emval_test_return_shared_ptr(engine, ctx)
			Expect(err).To(BeNil())

			_, err = generated.TakesHeldBySmartPtr(engine, ctx, e)
			Expect(err).To(Not(BeNil()))
			Expect(err.Error()).To(ContainSubstring("expected null or instance of HeldBySmartPtr, got an instance of ValHolder"))

			err = e.DeleteInstance(ctx, e)
			Expect(err).To(BeNil())
		})

		It("constructor can return smart pointer", func() {
			e, err := generated.NewClassHeldBySmartPtr(engine, ctx, 10, "foo")
			Expect(err).To(BeNil())

			i, err := e.GetPropertyI(ctx)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(int32(10)))

			s, err := e.GetPropertyS(ctx)
			Expect(err).To(BeNil())
			Expect(s).To(Equal("foo"))

			f, err := generated.TakesHeldBySmartPtr(engine, ctx, e)

			err = f.DeleteInstance(ctx, f)
			Expect(err).To(BeNil())

			err = e.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("custom smart pointer", func() {
			e, err := generated.NewClassHeldByCustomSmartPtr(engine, ctx, 20, "bar")
			Expect(err).To(BeNil())

			i, err := e.GetPropertyI(ctx)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(int32(20)))

			s, err := e.GetPropertyS(ctx)
			Expect(err).To(BeNil())
			Expect(s).To(Equal("bar"))

			err = e.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("custom smart pointer passed through wiretype", func() {
			e, err := generated.NewClassHeldByCustomSmartPtr(engine, ctx, 20, "bar")
			Expect(err).To(BeNil())

			f, err := generated.PassThroughCustomSmartPtr(engine, ctx, e)
			Expect(err).To(BeNil())

			err = e.Delete(ctx)
			Expect(err).To(BeNil())

			i, err := f.GetInstanceProperty(ctx, f, "i")
			Expect(err).To(BeNil())
			Expect(i).To(Equal(int32(20)))

			s, err := f.GetInstanceProperty(ctx, f, "s")
			Expect(err).To(BeNil())
			Expect(s).To(Equal("bar"))

			err = f.DeleteInstance(ctx, f)
			Expect(err).To(BeNil())
		})

		It("cannot give null to by-value argument", func() {
			e, err := generated.TakesHeldBySmartPtr(engine, ctx, nil)
			Expect(err).To(Not(BeNil()))
			Expect(err.Error()).To(ContainSubstring("nil is not a valid HeldBySmartPtr"))
			Expect(e).To(BeNil())
		})

		It("raw pointer can take and give null", func() {
			e, err := generated.PassThroughRawPtr(engine, ctx, nil)
			Expect(err).To(BeNil())
			Expect(e).To(BeNil())
		})

		It("custom smart pointer can take and give null", func() {
			e, err := generated.PassThroughCustomSmartPtr(engine, ctx, nil)
			Expect(err).To(BeNil())
			Expect(e).To(BeNil())
		})

		It("cannot pass shared_ptr to CustomSmartPtr", func() {
			o, err := generated.ClassHeldByCustomSmartPtrStaticCreateSharedPtr(engine, ctx, 10, "foo")
			Expect(err).To(BeNil())

			e, err := generated.PassThroughCustomSmartPtr(engine, ctx, o)
			Expect(err).To(Not(BeNil()))
			Expect(err.Error()).To(ContainSubstring("cannot convert argument of type shared_ptr<HeldByCustomSmartPtr> to parameter type CustomSmartPtr<HeldByCustomSmartPtr>"))
			Expect(e).To(BeNil())

			err = o.DeleteInstance(ctx, o)
			Expect(err).To(BeNil())
		})

		It("custom smart pointers can be passed to shared_ptr parameter", func() {
			e, err := generated.ClassHeldBySmartPtrStaticNewCustomPtr(engine, ctx, 10, "abc")
			Expect(err).To(BeNil())

			i, err := e.GetInstanceProperty(ctx, e, "i")
			Expect(err).To(BeNil())
			Expect(i, 10)

			s, err := e.GetInstanceProperty(ctx, e, "s")
			Expect(err).To(BeNil())
			Expect(s, "abc")

			tmp, err := generated.TakesHeldBySmartPtrSharedPtr(engine, ctx, e)
			Expect(err).To(BeNil())

			err = tmp.DeleteInstance(ctx, tmp)
			Expect(err).To(BeNil())

			err = e.DeleteInstance(ctx, e)
			Expect(err).To(BeNil())
		})

		It("can call non-member functions as methods", func() {
			e, err := generated.NewClassHeldBySmartPtr(engine, ctx, 20, "bar")
			Expect(err).To(BeNil())

			f, err := e.ReturnThis(ctx)
			Expect(err).To(BeNil())

			err = e.Delete(ctx)
			Expect(err).To(BeNil())

			i, err := f.GetInstanceProperty(ctx, f, "i")
			Expect(err).To(BeNil())
			Expect(i, 10)

			s, err := f.GetInstanceProperty(ctx, f, "s")
			Expect(err).To(BeNil())
			Expect(s, "abc")

			err = f.DeleteInstance(ctx, f)
			Expect(err).To(BeNil())
		})
	})

	When("enumerations", func() {
		It("can pass and return enumeration values to functions", func() {
			tar, err := generated.Emval_test_take_and_return_Enum(engine, ctx, generated.EnumEnum_TWO)
			Expect(err).To(BeNil())
			Expect(tar).To(Equal(generated.EnumEnum_TWO))
		})
	})

	When("C++11 enum class", func() {
		It("can pass and return enumeration values to functions", func() {
			tar, err := generated.Emval_test_take_and_return_EnumClass(engine, ctx, generated.EnumEnumClass_TWO)
			Expect(err).To(BeNil())
			Expect(tar).To(Equal(generated.EnumEnumClass_TWO))
		})
	})

	When("emval call tests", func() {
		It("can call functions from C++", func() {
			called := false
			err := generated.Emval_test_call_function(engine, ctx, func(i int32, f float32, tv []any, sv map[string]any) {
				called = true

				Expect(i).To(Equal(int32(10)))
				Expect(f).To(Equal(float32(1.5)))
				Expect(tv).To(Equal([]any{float32(1.25), float32(2.5), float32(3.75), float32(4)}))
				Expect(sv).To(Equal(map[string]any{"x": float32(1.25), "y": float32(2.5), "z": float32(3.75), "w": float32(4)}))
			}, 10, 1.5, []any{float32(1.25), float32(2.5), float32(3.75), float32(4)}, map[string]any{"x": float32(1.25), "y": float32(2.5), "z": float32(3.75), "w": float32(4)})
			Expect(err).To(BeNil())
			Expect(called).To(BeTrue())
		})
	})

	When("raw pointers", func() {
		It("can pass raw pointers into functions if explicitly allowed", func() {
			vh, err := generated.NewClassValHolder(engine, ctx, "foo")
			Expect(err).To(BeNil())

			err = vh.StaticSet_via_raw_pointer(ctx, vh, 10)
			Expect(err).To(BeNil())

			val, err := vh.StaticGet_via_raw_pointer(ctx, vh)
			Expect(err).To(BeNil())
			Expect(val).To(Equal(10))

			err = vh.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("can return raw pointers from functions if explicitly allowed", func() {
			p, err := generated.Embind_test_return_raw_base_ptr(engine, ctx)
			Expect(err).To(BeNil())

			className, err := p.CallInstanceMethod(ctx, p, "getClassName")
			Expect(className).To(Equal("Base"))

			err = p.DeleteInstance(ctx, p)
			Expect(err).To(BeNil())
		})

		It("can pass multiple raw pointers to functions", func() {
			target, err := generated.NewClassValHolder(engine, ctx, types.Undefined)
			Expect(err).To(BeNil())

			source, err := generated.NewClassValHolder(engine, ctx, "hi")
			Expect(err).To(BeNil())

			err = generated.ClassValHolderStaticTransfer_via_raw_pointer(engine, ctx, target, source)
			Expect(err).To(BeNil())

			val, err := target.GetVal(ctx)
			Expect(err).To(BeNil())

			Expect(val).To(Equal("hi"))

			err = target.Delete(ctx)
			Expect(err).To(BeNil())

			err = source.Delete(ctx)
			Expect(err).To(BeNil())
		})
	})

	When("implementing abstract methods with JS objects", func() {
		/*
		   test("can call abstract methods", function() {
		       var obj = cm.getAbstractClass();
		       assert.equal("from concrete", obj.abstractMethod());
		       obj.delete();
		   });

		   test("can implement abstract methods in JavaScript", function() {
		       var expected = "my JS string";
		       function MyImplementation() {
		           this.rv = expected;
		       }
		       MyImplementation.prototype.abstractMethod = function() {
		           return this.rv;
		       };

		       var impl = cm.AbstractClass.implement(new MyImplementation);
		       assert.equal(expected, impl.abstractMethod());
		       assert.equal(expected, cm.callAbstractMethod(impl));
		       impl.delete();
		   });

		   test("can implement optional methods in JavaScript", function() {
		       var expected = "my JS string";
		       function MyImplementation() {
		           this.rv = expected;
		       }
		       MyImplementation.prototype.optionalMethod = function() {
		           return this.rv;
		       };

		       var impl = cm.AbstractClass.implement(new MyImplementation);
		       assert.equal(expected, cm.callOptionalMethod(impl, expected));
		       impl.delete();
		   });

		   test("if not implemented then optional method runs default", function() {
		       var impl = cm.AbstractClass.implement({});
		       assert.equal("optionalfoo", impl.optionalMethod("foo"));
		       impl.delete();
		   });

		   test("returning null shared pointer from interfaces implemented in JS code does not leak", function() {
		       var impl = cm.AbstractClass.implement({
		           returnsSharedPtr: function() {
		               return null;
		           }
		       });
		       cm.callReturnsSharedPtrMethod(impl);
		       impl.delete();
		       // Let the memory leak test superfixture check that no leaks occurred.
		   });

		   test("returning a new shared pointer from interfaces implemented in JS code does not leak", function() {
		       var impl = cm.AbstractClass.implement({
		           returnsSharedPtr: function() {
		               return cm.embind_test_return_smart_derived_ptr().deleteLater();
		           }
		       });
		       cm.callReturnsSharedPtrMethod(impl);
		       impl.delete();
		       // Let the memory leak test superfixture check that no leaks occurred.
		   });

		   test("void methods work", function() {
		       var saved = {};
		       var impl = cm.AbstractClass.implement({
		           differentArguments: function(i, d, f, q, s) {
		               saved.i = i;
		               saved.d = d;
		               saved.f = f;
		               saved.q = q;
		               saved.s = s;
		           }
		       });

		       cm.callDifferentArguments(impl, 1, 2, 3, 4, "foo");

		       assert.deepEqual(saved, {
		           i: 1,
		           d: 2,
		           f: 3,
		           q: 4,
		           s: "foo",
		       });

		       impl.delete();
		   });

		   test("returning a cached new shared pointer from interfaces implemented in JS code does not leak", function() {
		       var derived = cm.embind_test_return_smart_derived_ptr();
		       var impl = cm.AbstractClass.implement({
		           returnsSharedPtr: function() {
		               return derived;
		           }
		       });
		       cm.callReturnsSharedPtrMethod(impl);
		       impl.delete();
		       derived.delete();
		       // Let the memory leak test superfixture check that no leaks occurred.
		   });
		*/
	})

	When("constructor prototype class inheritance", func() {
		/*
		   var Empty = cm.AbstractClass.extend("Empty", {
		       abstractMethod: function() {
		       }
		   });

		   test("can extend, construct, and delete", function() {
		       var instance = new Empty;
		       instance.delete();
		   });

		   test("properties set in constructor are externally visible", function() {
		       var HasProperty = cm.AbstractClass.extend("HasProperty", {
		           __construct: function(x) {
		               this.__parent.__construct.call(this);
		               this.property = x;
		           },
		           abstractMethod: function() {
		           }
		       });
		       var instance = new HasProperty(10);
		       assert.equal(10, instance.property);
		       instance.delete();
		   });

		   test("pass derived object to c++", function() {
		       var Implementation = cm.AbstractClass.extend("Implementation", {
		           abstractMethod: function() {
		               return "abc";
		           },
		       });
		       var instance = new Implementation;
		       var result = cm.callAbstractMethod(instance);
		       instance.delete();
		       assert.equal("abc", result);
		   });

		   test("properties set in constructor are visible in overridden methods", function() {
		       var HasProperty = cm.AbstractClass.extend("HasProperty", {
		           __construct: function(x) {
		               this.__parent.__construct.call(this);
		               this.x = x;
		           },
		           abstractMethod: function() {
		               return this.x;
		           },
		       });
		       var instance = new HasProperty("xyz");
		       var result = cm.callAbstractMethod(instance);
		       instance.delete();
		       assert.equal("xyz", result);
		   });

		   test("interface methods are externally visible", function() {
		       var instance = new Empty;
		       var result = instance.concreteMethod();
		       instance.delete();
		       assert.equal("concrete", result);
		   });

		   test("optional methods are externally visible", function() {
		       var instance = new Empty;
		       var result = instance.optionalMethod("_123");
		       instance.delete();
		       assert.equal("optional_123", result);
		   });

		   test("optional methods: not defined", function() {
		       var instance = new Empty;
		       var result = cm.callOptionalMethod(instance, "_123");
		       instance.delete();
		       assert.equal("optional_123", result);
		   });

		   // Calling C++ implementations of optional functions can be
		   // made to work, but requires an interface change on the C++
		   // side, using a technique similar to the one described at
		   // https://wiki.python.org/moin/boost.python/OverridableVirtualFunctions
		   //
		   // The issue is that, in a standard binding, calling
		   // parent.prototype.optionalMethod invokes the wrapper
		   // function, which checks that the JS object implements
		   // 'optionalMethod', which it does.  Thus, C++ calls back into
		   // JS, resulting in an infinite loop.
		   //
		   // The solution, for optional methods, is to bind a special
		   // concrete implementation that specifically calls the base
		   // class's implementation.  See the binding of
		   // AbstractClass::optionalMethod in embind_test.cpp.

		   test("can call parent implementation from within derived implementation", function() {
		       var parent = cm.AbstractClass;
		       var ExtendsOptionalMethod = parent.extend("ExtendsOptionalMethod", {
		           abstractMethod: function() {
		           },
		           optionalMethod: function(s) {
		               return "optionaljs_" + parent.prototype.optionalMethod.call(this, s);
		           },
		       });
		       var instance = new ExtendsOptionalMethod;
		       var result = cm.callOptionalMethod(instance, "_123");
		       instance.delete();
		       assert.equal("optionaljs_optional_123", result);
		   });

		   test("instanceof", function() {
		       var instance = new Empty;
		       assert.instanceof(instance, Empty);
		       assert.instanceof(instance, cm.AbstractClass);
		       instance.delete();
		   });

		   test("returning null shared pointer from interfaces implemented in JS code does not leak", function() {
		       var C = cm.AbstractClass.extend("C", {
		           abstractMethod: function() {
		           },
		           returnsSharedPtr: function() {
		               return null;
		           }
		       });
		       var impl = new C;
		       cm.callReturnsSharedPtrMethod(impl);
		       impl.delete();
		       // Let the memory leak test superfixture check that no leaks occurred.
		   });

		   test("returning a new shared pointer from interfaces implemented in JS code does not leak", function() {
		       var C = cm.AbstractClass.extend("C", {
		           abstractMethod: function() {
		           },
		           returnsSharedPtr: function() {
		               return cm.embind_test_return_smart_derived_ptr().deleteLater();
		           }
		       });
		       var impl = new C;
		       cm.callReturnsSharedPtrMethod(impl);
		       impl.delete();
		       // Let the memory leak test superfixture check that no leaks occurred.
		   });

		   test("void methods work", function() {
		       var saved = {};
		       var C = cm.AbstractClass.extend("C", {
		           abstractMethod: function() {
		           },
		           differentArguments: function(i, d, f, q, s) {
		               saved.i = i;
		               saved.d = d;
		               saved.f = f;
		               saved.q = q;
		               saved.s = s;
		           }
		       });
		       var impl = new C;

		       cm.callDifferentArguments(impl, 1, 2, 3, 4, "foo");

		       assert.deepEqual(saved, {
		           i: 1,
		           d: 2,
		           f: 3,
		           q: 4,
		           s: "foo",
		       });

		       impl.delete();
		   });

		   test("returning a cached new shared pointer from interfaces implemented in JS code does not leak", function() {
		       var derived = cm.embind_test_return_smart_derived_ptr();
		       var C = cm.AbstractClass.extend("C", {
		           abstractMethod: function() {
		           },
		           returnsSharedPtr: function() {
		               return derived;
		           }
		       });
		       var impl = new C;
		       cm.callReturnsSharedPtrMethod(impl);
		       impl.delete();
		       derived.delete();
		       // Let the memory leak test superfixture check that no leaks occurred.
		   });

		   test("calling pure virtual function gives good error message", function() {
		       var C = cm.AbstractClass.extend("C", {});
		       var error = assert.throws(cm.PureVirtualError, function() {
		           new C;
		       });
		       assert.equal('Pure virtual function abstractMethod must be implemented in JavaScript', error.message);
		   });

		   test("can extend from C++ class with constructor arguments", function() {
		       var parent = cm.AbstractClassWithConstructor;
		       var C = parent.extend("C", {
		           __construct: function(x) {
		               this.__parent.__construct.call(this, x);
		           },
		           abstractMethod: function() {
		               return this.concreteMethod();
		           }
		       });

		       var impl = new C("hi");
		       var rv = cm.callAbstractMethod2(impl);
		       impl.delete();

		       assert.equal("hi", rv);
		   });

		   test("__destruct is called when object is destroyed", function() {
		       var parent = cm.HeldAbstractClass;
		       var calls = [];
		       var C = parent.extend("C", {
		           method: function() {
		           },
		           __destruct: function() {
		               calls.push("__destruct");
		               this.__parent.__destruct.call(this);
		           }
		       });
		       var impl = new C;
		       var copy = impl.clone();
		       impl.delete();
		       assert.deepEqual([], calls);
		       copy.delete();
		       assert.deepEqual(["__destruct"], calls);
		   });

		   test("if JavaScript implementation of interface is returned, don't wrap in new handle", function() {
		       var parent = cm.HeldAbstractClass;
		       var C = parent.extend("C", {
		           method: function() {
		           }
		       });
		       var impl = new C;
		       var rv = cm.passHeldAbstractClass(impl);
		       impl.delete();
		       assert.equal(impl, rv);
		       rv.delete();
		   });

		   test("can instantiate two wrappers with constructors", function() {
		       var parent = cm.HeldAbstractClass;
		       var C = parent.extend("C", {
		           __construct: function() {
		               this.__parent.__construct.call(this);
		           },
		           method: function() {
		           }
		       });
		       var a = new C;
		       var b = new C;
		       a.delete();
		       b.delete();
		   });

		   test("incorrectly calling parent is an error", function() {
		       var parent = cm.HeldAbstractClass;
		       var C = parent.extend("C", {
		           __construct: function() {
		               this.__parent.__construct();
		           },
		           method: function() {
		           }
		       });
		       assert.throws(cm.BindingError, function() {
		           new C;
		       });
		   });

		   test("deleteLater() works for JavaScript implementations", function() {
		       var parent = cm.HeldAbstractClass;
		       var C = parent.extend("C", {
		           method: function() {
		           }
		       });
		       var impl = new C;
		       var rv = cm.passHeldAbstractClass(impl);
		       impl.deleteLater();
		       rv.deleteLater();
		       cm.flushPendingDeletes();
		   });

		   test("deleteLater() combined with delete() works for JavaScript implementations", function() {
		       var parent = cm.HeldAbstractClass;
		       var C = parent.extend("C", {
		           method: function() {
		           }
		       });
		       var impl = new C;
		       var rv = cm.passHeldAbstractClass(impl);
		       impl.deleteLater();
		       rv.delete();
		       cm.flushPendingDeletes();
		   });

		   test("method arguments with pointer ownership semantics are cleaned up after call", function() {
		       var parent = cm.AbstractClass;
		       var C = parent.extend("C", {
		           abstractMethod: function() {
		           },
		       });
		       var impl = new C;
		       cm.passShared(impl);
		       impl.delete();
		   });

		   test("method arguments with pointer ownership semantics can be cloned", function() {
		       var parent = cm.AbstractClass;
		       var owned;
		       var C = parent.extend("C", {
		           abstractMethod: function() {
		           },
		           passShared: function(p) {
		               owned = p.clone();
		           }
		       });
		       var impl = new C;
		       cm.passShared(impl);
		       impl.delete();

		       assert.equal("Derived", owned.getClassName());
		       owned.delete();
		   });
		*/

		It("emscripten::val method arguments don't leak", func() {
			type newStructTypeToExtend struct {
				embind_external.ClassBase
			}
			parent, err := generated.ClassAbstractClassStaticExtend(engine, ctx, "C2", &newStructTypeToExtend{})
			Expect(err).To(BeNil())

			impl, err := parent.(func(context.Context, ...any) (any, error))(ctx)
			Expect(err).To(BeNil())

			typedImpl := impl.(*newStructTypeToExtend)
			err = typedImpl.DeleteInstance(ctx, typedImpl)
			Expect(err).To(BeNil())

			// @todo: do we actually want to test whether this works?
			//   We already know this of the correct type, so why do we need to validate this?
			//   Go already tells us this is a specific type.
		})
	})

	When("registration order", func() {
		It("registration of tuple elements out of order leaves them in order", func() {
			ot, err := generated.GetOrderedTuple(engine, ctx)
			Expect(err).To(BeNil())

			Expect(ot).To(HaveLen(2))
			Expect(ot[0]).To(BeAssignableToTypeOf(&generated.ClassFirstElement{}))
			Expect(ot[1]).To(BeAssignableToTypeOf(&generated.ClassSecondElement{}))

			err = ot[0].(*generated.ClassFirstElement).Delete(ctx)
			Expect(err).To(BeNil())

			err = ot[1].(*generated.ClassSecondElement).Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("registration of struct elements out of order", func() {
			ot, err := generated.GetOrderedStruct(engine, ctx)
			Expect(err).To(BeNil())

			Expect(ot).To(HaveKey("first"))
			Expect(ot).To(HaveKey("second"))

			Expect(ot["first"]).To(BeAssignableToTypeOf(&generated.ClassFirstElement{}))
			Expect(ot["second"]).To(BeAssignableToTypeOf(&generated.ClassSecondElement{}))

			err = ot["first"].(*generated.ClassFirstElement).Delete(ctx)
			Expect(err).To(BeNil())

			err = ot["second"].(*generated.ClassSecondElement).Delete(ctx)
			Expect(err).To(BeNil())
		})
	})

	When("unbound types", func() {
		if !generated.Constant_hasUnboundTypeNames {
			return
		}

		assertMessage := func(fn func() (any, error), message string) {
			_, err := fn()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring(message))
		}

		It("calling function with unbound types produces error", func() {
			assertMessage(
				func() (any, error) {
					return generated.GetUnboundClass(engine, ctx, 1)
				},
				"getUnboundClass due to unbound types: 12UnboundClass")
		})

		It("unbound base class produces error", func() {
			assertMessage(
				func() (any, error) {
					return generated.GetHasUnboundBase(engine, ctx, 1)
				},
				"getHasUnboundBase due to unbound types: 12UnboundClass")
		})

		It("construct of class with unbound base", func() {
			assertMessage(
				func() (any, error) {
					return generated.HasUnboundBase(engine, ctx)
				}, "HasUnboundBase due to unbound types: 12UnboundClass")
		})

		It("unbound constructor argument", func() {
			assertMessage(
				func() (any, error) {
					return generated.NewClassHasConstructorUsingUnboundArgument(engine, ctx, 1)
				},
				"HasConstructorUsingUnboundArgument due to unbound types: 12UnboundClass")
		})

		It("unbound constructor argument of class with unbound base", func() {
			assertMessage(
				func() (any, error) {
					return generated.HasConstructorUsingUnboundArgumentAndUnboundBase(engine, ctx)
				},
				"HasConstructorUsingUnboundArgumentAndUnboundBase due to unbound types: 18SecondUnboundClass")
		})

		It("class function with unbound argument", func() {
			x, err := generated.NewClassBoundClass(engine, ctx)
			Expect(err).To(BeNil())

			assertMessage(
				func() (any, error) {
					return x.Method(ctx, 0)
				}, "Cannot call BoundClass.method due to unbound types: 12UnboundClass")
			err = x.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("class class function with unbound argument", func() {
			assertMessage(
				func() (any, error) {
					return generated.ClassBoundClassStaticClassfunction(engine, ctx, 0)
				}, "Cannot call BoundClass.classfunction due to unbound types: 12UnboundClass")
		})

		It("class property of unbound type", func() {
			x, err := generated.NewClassBoundClass(engine, ctx)
			Expect(err).To(BeNil())

			assertMessage(
				func() (any, error) {
					return x.GetPropertyProperty(ctx)
				}, "Cannot access BoundClass.property due to unbound types: 12UnboundClass")
			assertMessage(
				func() (any, error) {
					err := x.SetPropertyProperty(ctx, 10)
					return nil, err
				}, "Cannot access BoundClass.property due to unbound types: 12UnboundClass")
			err = x.Delete(ctx)
			Expect(err).To(BeNil())
		})
	})

	When("noncopyable", func() {
		It("can call method on noncopyable object", func() {
			x, err := generated.NewClassNoncopyable(engine, ctx)
			Expect(err).To(BeNil())

			method, err := x.Method(ctx)
			Expect(err).To(BeNil())
			Expect(method).To(Equal("foo"))

			err = x.Delete(ctx)
			Expect(err).To(BeNil())
		})
	})

	When("constants", func() {
		Expect(generated.Constant_INT_CONSTANT, 10)
		Expect(generated.Constant_STATIC_CONST_INTEGER_VALUE_1, 1)
		Expect(generated.Constant_STATIC_CONST_INTEGER_VALUE_1000, 1000)
		Expect(generated.Constant_STRING_CONSTANT, "some string")
		Expect(generated.Constant_VALUE_ARRAY_CONSTANT, []any{float32(1), float32(2), float32(3), float32(4)})
		Expect(generated.Constant_VALUE_OBJECT_CONSTANT, map[string]interface{}{"w": 4, "x": 1, "y": 2, "z": 3})
	})

	When("object handle comparison", func() {
		It("", func() {
			e, err := generated.NewClassValHolder(engine, ctx, "foo")
			Expect(err).To(BeNil())

			f, err := generated.NewClassValHolder(engine, ctx, "foo")
			Expect(err).To(BeNil())

			eIsAliasOfE, err := e.IsAliasOf(ctx, e)
			Expect(err).To(BeNil())
			Expect(eIsAliasOfE).To(BeTrue())

			eIsAliasOfF, err := e.IsAliasOf(ctx, f)
			Expect(err).To(BeNil())
			Expect(eIsAliasOfF).To(BeFalse())

			fIsAliasOfE, err := f.IsAliasOf(ctx, e)
			Expect(err).To(BeNil())
			Expect(fIsAliasOfE).To(BeFalse())

			err = e.Delete(ctx)
			Expect(err).To(BeNil())

			err = f.Delete(ctx)
			Expect(err).To(BeNil())
		})
	})

	When("derived-with-offset types compare with base", func() {
		It("", func() {
			e, err := generated.NewClassDerivedWithOffset(engine, ctx)
			Expect(err).To(BeNil())

			f, err := generated.Return_Base_from_DerivedWithOffset(engine, ctx, e)
			Expect(err).To(BeNil())

			eIsAliasOfF, err := e.IsAliasOf(ctx, f)
			Expect(err).To(BeNil())
			Expect(eIsAliasOfF).To(BeTrue())

			fIsAliasOfE, err := f.(*generated.ClassBase).IsAliasOf(ctx, e)
			Expect(err).To(BeNil())
			Expect(fIsAliasOfE).To(BeTrue())

			err = e.Delete(ctx)
			Expect(err).To(BeNil())

			err = f.(*generated.ClassBase).Delete(ctx)
			Expect(err).To(BeNil())
		})
	})

	When("memory view", func() {

		It("can pass memory view from C++ to JS", func() {
			views := []any{}
			err := generated.CallWithMemoryView(engine, ctx, func(view any) {
				views = append(views, view)
			})
			Expect(err).To(BeNil())
			Expect(views).To(HaveLen(3))

			Expect(views[0]).To(HaveLen(8))
			Expect(views[0]).To(Equal([]uint8{0, 1, 2, 3, 4, 5, 6, 7}))

			Expect(views[1]).To(HaveLen(4))
			Expect(views[1]).To(Equal([]float32{1.5, 2.5, 3.5, 4.5}))

			Expect(views[2]).To(HaveLen(4))
			Expect(views[2]).To(Equal([]int16{1000, 100, 10, 1}))
		})
	})

	When("delete pool", func() {
		It("can delete objects later", func() {
			v, err := generated.NewClassValHolder(engine, ctx, struct{}{})
			Expect(err).To(BeNil())

			_, err = v.DeleteLater(ctx)
			Expect(err).To(BeNil())

			val, err := v.GetVal(ctx)
			Expect(err).To(BeNil())
			Expect(val).To(Equal(struct{}{}))

			err = engine.FlushPendingDeletes(ctx)

			val, err = v.GetVal(ctx)
			Expect(err).To(Not(BeNil()))
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("cannot pass deleted object as a pointer of type"))
			}
		})

		It("calling deleteLater twice is an error", func() {
			v, err := generated.NewClassValHolder(engine, ctx, struct{}{})
			Expect(err).To(BeNil())

			_, err = v.DeleteLater(ctx)
			Expect(err).To(BeNil())

			_, err = v.DeleteLater(ctx)
			Expect(err).To(Not(BeNil()))
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("object already scheduled for deletion"))
			}
		})

		It("can clone instances that have been scheduled for deletion", func() {
			v, err := generated.NewClassValHolder(engine, ctx, struct{}{})
			Expect(err).To(BeNil())

			_, err = v.DeleteLater(ctx)
			Expect(err).To(BeNil())

			v2, err := v.Clone(ctx)
			Expect(err).To(BeNil())

			err = v2.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("deleteLater returns the object", func() {
			v, err := generated.NewClassValHolder(engine, ctx, struct{}{})
			Expect(err).To(BeNil())

			vReturned, err := v.DeleteLater(ctx)
			Expect(err).To(BeNil())

			val, err := vReturned.(*generated.ClassValHolder).GetVal(ctx)
			Expect(err).To(BeNil())
			Expect(val).To(Equal(struct{}{}))
		})

		It("deleteLater throws if object is already deleted", func() {
			v, err := generated.NewClassValHolder(engine, ctx, struct{}{})
			Expect(err).To(BeNil())

			err = v.Delete(ctx)
			Expect(err).To(BeNil())

			_, err = v.DeleteLater(ctx)
			Expect(err).To(Not(BeNil()))
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("class handle already deleted"))
			}
		})

		It("delete throws if object is already scheduled for deletion", func() {
			v, err := generated.NewClassValHolder(engine, ctx, struct{}{})
			Expect(err).To(BeNil())

			_, err = v.DeleteLater(ctx)
			Expect(err).To(BeNil())

			err = v.Delete(ctx)
			Expect(err).To(Not(BeNil()))
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("object already scheduled for deletion"))
			}
		})

		It("deleteLater invokes delay function", func() {
			var runLater func(ctx context.Context) error

			err := engine.SetDelayFunction(func(fn func(ctx context.Context) error) error {
				runLater = fn
				return nil
			})

			Expect(err).To(BeNil())

			v, err := generated.NewClassValHolder(engine, ctx, struct{}{})
			Expect(err).To(BeNil())

			Expect(runLater).To(BeNil())

			_, err = v.DeleteLater(ctx)
			Expect(err).To(BeNil())

			Expect(runLater).To(Not(BeNil()))

			isDeleted := v.IsDeleted(ctx)

			Expect(isDeleted).To(BeFalse())
			err = runLater(ctx)
			Expect(err).To(BeNil())

			isDeleted = v.IsDeleted(ctx)
			Expect(isDeleted).To(BeTrue())
		})

		It("deleteLater twice invokes delay function once", func() {
			count := 0
			var runLater func(ctx context.Context) error

			err := engine.SetDelayFunction(func(fn func(ctx context.Context) error) error {
				count++
				runLater = fn
				return nil
			})

			Expect(err).To(BeNil())

			v, err := generated.NewClassValHolder(engine, ctx, struct{}{})
			Expect(err).To(BeNil())

			_, err = v.DeleteLater(ctx)
			Expect(err).To(BeNil())

			v2, err := generated.NewClassValHolder(engine, ctx, struct{}{})
			Expect(err).To(BeNil())

			_, err = v2.DeleteLater(ctx)
			Expect(err).To(BeNil())

			Expect(count).To(Equal(1))

			err = runLater(ctx)
			Expect(err).To(BeNil())

			v3, err := generated.NewClassValHolder(engine, ctx, struct{}{})
			Expect(err).To(BeNil())

			_, err = v3.DeleteLater(ctx)
			Expect(err).To(BeNil())

			Expect(count).To(Equal(2))
		})

		It("The delay function is immediately invoked if the deletion queue is not empty", func() {
			v, err := generated.NewClassValHolder(engine, ctx, struct{}{})
			Expect(err).To(BeNil())

			_, err = v.DeleteLater(ctx)
			Expect(err).To(BeNil())

			count := 0
			err = engine.SetDelayFunction(func(fn func(ctx context.Context) error) error {
				count++
				return nil
			})
			Expect(err).To(BeNil())
			Expect(count).To(Equal(1))
		})

		// The idea is that an interactive application would
		// periodically flush the deleteLater queue by calling
		//
		// setDelayFunction(function(fn) {
		//     setTimeout(fn, 0);
		// });
	})

	When("references", func() {
		It("JS object handles can be passed through to C++ by reference", func() {
			sh1, err := generated.NewClassStringHolder(engine, ctx, "Hello world")
			Expect(err).To(BeNil())

			sh1String, err := sh1.Get(ctx)
			Expect(err).To(BeNil())
			Expect(sh1String).To(Equal("Hello world"))

			err = generated.Clear_StringHolder(engine, ctx, sh1)
			Expect(err).To(BeNil())

			sh1String, err = sh1.Get(ctx)
			Expect(err).To(BeNil())
			Expect(sh1String).To(Equal(""))

			err = sh1.Delete(ctx)
			Expect(err).To(BeNil())
		})
	})

	When("val::as from pointer to value", func() {
		It("calling as on pointer with value makes a copy", func() {
			sh1, err := generated.NewClassStringHolder(engine, ctx, "Hello world")
			Expect(err).To(BeNil())

			sh2, err := generated.Return_StringHolder_copy(engine, ctx, sh1)
			Expect(err).To(BeNil())

			isAliasOf, err := sh1.IsAliasOf(ctx, sh2)
			Expect(err).To(BeNil())
			Expect(isAliasOf).To(BeFalse())

			err = sh2.(*generated.ClassStringHolder).Delete(ctx)
			Expect(err).To(BeNil())

			err = sh1.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("calling function that returns a StringHolder", func() {
			sh1, err := generated.NewClassStringHolder(engine, ctx, "Hello world")
			Expect(err).To(BeNil())

			sh2, err := generated.Call_StringHolder_func(engine, ctx, func() *generated.ClassStringHolder {
				return sh1
			})
			Expect(err).To(BeNil())

			sh1String, err := sh1.Get(ctx)
			Expect(err).To(BeNil())
			Expect(sh1String).To(Equal("Hello world"))

			sh2String, err := sh2.(*generated.ClassStringHolder).Get(ctx)
			Expect(err).To(BeNil())
			Expect(sh2String).To(Equal("Hello world"))

			isAliasOf, err := sh1.IsAliasOf(ctx, sh2)
			Expect(err).To(BeNil())
			Expect(isAliasOf).To(BeFalse())

			err = sh2.(*generated.ClassStringHolder).Delete(ctx)
			Expect(err).To(BeNil())

			err = sh1.Delete(ctx)
			Expect(err).To(BeNil())
		})
	})

	When("mixin", func() {
		It("can call mixin method", func() {
			a, err := generated.NewClassDerivedWithMixin(engine, ctx)
			Expect(err).To(BeNil())

			// assert.instanceof(a, cm.Base);

			get10, err := a.Get10(ctx)
			Expect(err).To(BeNil())
			Expect(get10).To(Equal(int32(10)))

			err = a.Delete(ctx)
			Expect(err).To(BeNil())
		})
	})

	When("val::as", func() {
		It("built-ins", func() {
			valAsBool, err := generated.Val_as_bool(engine, ctx, true)
			Expect(err).To(BeNil())
			Expect(valAsBool).To(BeTrue())

			valAsBool, err = generated.Val_as_bool(engine, ctx, false)
			Expect(err).To(BeNil())
			Expect(valAsBool).To(BeFalse())

			valAsChar, err := generated.Val_as_char(engine, ctx, int8(127))
			Expect(err).To(BeNil())
			Expect(valAsChar).To(Equal(int8(127)))

			valAsShort, err := generated.Val_as_short(engine, ctx, int16(32767))
			Expect(err).To(BeNil())
			Expect(valAsShort).To(Equal(int16(32767)))

			valAsInt, err := generated.Val_as_int(engine, ctx, int32(65536))
			Expect(err).To(BeNil())
			Expect(valAsInt).To(Equal(int32(65536)))

			valAsLong, err := generated.Val_as_long(engine, ctx, int32(65536))
			Expect(err).To(BeNil())
			Expect(valAsLong).To(Equal(int32(65536)))

			valAsDouble, err := generated.Val_as_float(engine, ctx, float32(10.5))
			Expect(err).To(BeNil())
			Expect(valAsDouble).To(Equal(float32(10.5)))

			valAsFloat, err := generated.Val_as_double(engine, ctx, float64(10.5))
			Expect(err).To(BeNil())
			Expect(valAsFloat).To(Equal(float64(10.5)))

			valAsString, err := generated.Val_as_string(engine, ctx, "foo")
			Expect(err).To(BeNil())
			Expect(valAsString).To(Equal("foo"))

			valAsWString, err := generated.Val_as_wstring(engine, ctx, "foo")
			Expect(err).To(BeNil())
			Expect(valAsWString).To(Equal("foo"))

			obj := struct{}{}

			valAsObj, err := generated.Val_as_val(engine, ctx, obj)
			Expect(err).To(BeNil())
			Expect(valAsObj).To(Equal(obj))

			// JS->C++ memory view not implemented (comment from emscripten)
			//var ab = cm.val_as_memory_view(new ArrayBuffer(13));
			//assert.equal(13, ab.byteLength);
		})

		It("value types", func() {
			tuple := []any{float32(1), float32(2), float32(3), float32(4)}

			valAsValueArray, err := generated.Val_as_value_array(engine, ctx, tuple)
			Expect(err).To(BeNil())
			Expect(valAsValueArray).To(Equal(tuple))

			valStruct := map[string]any{
				"x": float32(1),
				"y": float32(2),
				"z": float32(3),
				"w": float32(4),
			}

			valAsValueStruct, err := generated.Val_as_value_object(engine, ctx, valStruct)
			Expect(err).To(BeNil())
			Expect(valAsValueStruct).To(Equal(valStruct))
		})

		It("enums", func() {
			valAsEnum, err := generated.Val_as_enum(engine, ctx, generated.EnumEnum_ONE)
			Expect(err).To(BeNil())
			Expect(valAsEnum).To(Equal(generated.EnumEnum_ONE))
		})
	})

	When("val::new_", func() {
		It("variety of types", func() {
			type factoryStruct struct {
				Arg1 uint8                   `embind_arg:"0"`
				Arg2 float64                 `embind_arg:"1"`
				Arg3 string                  `embind_arg:"2"`
				Arg4 map[string]any          `embind_arg:"3"`
				Arg5 generated.EnumEnumClass `embind_arg:"4"`
				Arg6 []any                   `embind_arg:"5"`
			}
			instance, err := generated.Construct_with_6_arguments(engine, ctx, factoryStruct{})
			Expect(err).To(BeNil())
			Expect(instance).To(Equal(factoryStruct{
				Arg1: 6,
				Arg2: -12.5,
				Arg3: "a3",
				Arg4: map[string]any{
					"x": float32(1),
					"y": float32(2),
					"z": float32(3),
					"w": float32(4),
				},
				Arg5: generated.EnumEnumClass_TWO,
				Arg6: []any{
					float32(-1),
					float32(-2),
					float32(-3),
					float32(-4),
				},
			}))
		})

		It("memory view", func() {
			type factoryStruct struct {
				Before string `embind_arg:"0"`
				View   []int8 `embind_arg:"1"`
				After  string `embind_arg:"2"`
			}

			instance, err := generated.Construct_with_memory_view(engine, ctx, factoryStruct{})
			Expect(err).To(BeNil())
			Expect(instance).To(Equal(factoryStruct{
				Before: "before",
				View:   []int8{48, 49, 50, 51, 52, 53, 54, 55, 56, 57},
				After:  "after",
			}))
		})

		It("ints_and_float", func() {
			type factoryStruct struct {
				A int32   `embind_arg:"0"`
				B float32 `embind_arg:"1"`
				C int32   `embind_arg:"2"`
			}

			instance, err := generated.Construct_with_ints_and_float(engine, ctx, factoryStruct{})
			Expect(err).To(BeNil())
			Expect(instance).To(Equal(factoryStruct{
				A: 65537,
				B: 4.0,
				C: 65538,
			}))
		})
	})

	When("intrusive pointers", func() {
		It("can pass intrusive pointers", func() {
			ic, err := generated.NewClassIntrusiveClass(engine, ctx)
			Expect(err).To(BeNil())

			d, err := generated.PassThroughIntrusiveClass(engine, ctx, ic)
			Expect(err).To(BeNil())

			isAlias, err := ic.IsAliasOf(ctx, d)
			Expect(err).To(BeNil())
			Expect(isAlias).To(Equal(true))

			err = ic.Delete(ctx)
			Expect(err).To(BeNil())

			err = d.(*generated.ClassIntrusiveClass).Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("can hold intrusive pointers", func() {
			ic, err := generated.NewClassIntrusiveClass(engine, ctx)
			Expect(err).To(BeNil())

			holder, err := generated.NewClassIntrusiveClassHolder(engine, ctx)
			Expect(err).To(BeNil())

			err = holder.Set(ctx, ic)
			Expect(err).To(BeNil())

			err = ic.Delete(ctx)
			Expect(err).To(BeNil())

			d, err := holder.Get(ctx)
			Expect(err).To(BeNil())

			err = d.(*generated.ClassIntrusiveClass).Delete(ctx)
			Expect(err).To(BeNil())

			err = holder.Delete(ctx)
			Expect(err).To(BeNil())
		})

		It("can extend from intrusive pointer class and still preserve reference in JavaScript", func() {
			type newStructTypeToExtend struct {
				embind_external.ClassBase
			}
			C, err := generated.ClassIntrusiveClassStaticExtend(engine, ctx, "C2", &newStructTypeToExtend{})
			Expect(err).To(BeNil())

			instance, err := C.(func(context.Context, ...any) (any, error))(ctx)
			Expect(err).To(BeNil())

			typedInstance := instance.(embind_external.ClassBase)

			holder, err := generated.NewClassIntrusiveClassHolder(engine, ctx)
			Expect(err).To(BeNil())

			err = holder.Set(ctx, typedInstance)
			Expect(err).To(BeNil())

			err = typedInstance.DeleteInstance(ctx, typedInstance)
			Expect(err).To(BeNil())

			back, err := holder.Get(ctx)
			Expect(err).To(BeNil())

			Expect(back).To(Equal(instance))

			err = holder.Delete(ctx)
			Expect(err).To(BeNil())

			err = back.DeleteInstance(ctx, back)
			Expect(err).To(BeNil())
		})
	})

	When("typeof", func() {
		It("typeof", func() {
			typeName, err := generated.GetTypeOfVal(engine, ctx, nil)
			Expect(err).To(BeNil())
			Expect(typeName).To(Equal("object"))

			typeName, err = generated.GetTypeOfVal(engine, ctx, struct{}{})
			Expect(err).To(BeNil())
			Expect(typeName).To(Equal("object"))

			typeName, err = generated.GetTypeOfVal(engine, ctx, func() {})
			Expect(err).To(BeNil())
			Expect(typeName).To(Equal("function"))

			typeName, err = generated.GetTypeOfVal(engine, ctx, 1)
			Expect(err).To(BeNil())
			Expect(typeName).To(Equal("number"))

			typeName, err = generated.GetTypeOfVal(engine, ctx, "hi")
			Expect(err).To(BeNil())
			Expect(typeName).To(Equal("string"))

			typeName, err = generated.GetTypeOfVal(engine, ctx, types.Undefined)
			Expect(err).To(BeNil())
			Expect(typeName).To(Equal("undefined"))

			typeName, err = generated.GetTypeOfVal(engine, ctx, true)
			Expect(err).To(BeNil())
			Expect(typeName).To(Equal("boolean"))

			typeName, err = generated.GetTypeOfVal(engine, ctx, int64(0))
			Expect(err).To(BeNil())
			Expect(typeName).To(Equal("bigint"))

			typeName, err = generated.GetTypeOfVal(engine, ctx, uint64(0))
			Expect(err).To(BeNil())
			Expect(typeName).To(Equal("bigint"))
		})
	})

	When("static members", func() {
		It("static members", func() {
			c, err := generated.ClassHasStaticMemberGetStaticPropertyC(engine, ctx)
			Expect(err).To(BeNil())
			Expect(c).To(Equal(int32(10)))

			v, err := generated.ClassHasStaticMemberGetStaticPropertyV(engine, ctx)
			Expect(err).To(BeNil())
			Expect(v).To(Equal(int32(20)))

			err = generated.ClassHasStaticMemberSetStaticPropertyV(engine, ctx, 30)
			Expect(err).To(BeNil())

			v, err = generated.ClassHasStaticMemberGetStaticPropertyV(engine, ctx)
			Expect(err).To(BeNil())
			Expect(v).To(Equal(int32(30)))
		})
	})
})
