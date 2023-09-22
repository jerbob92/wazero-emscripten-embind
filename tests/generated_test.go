package tests

import (
	"context"
	"log"
	"os"
	"testing"

	embind_external "github.com/jerbob92/wazero-emscripten-embind"
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
		/*
					        test("value creation", function() {
					            assert.equal(15, cm.emval_test_new_integer());
					            assert.equal("Hello everyone", cm.emval_test_new_string());
					            assert.equal("Hello everyone", cm.emval_test_get_string_from_val({key: "Hello everyone"}));

					            var object = cm.emval_test_new_object();
					            assert.equal('bar', object.foo);
					            assert.equal(1, object.baz);
					        });

					        test("pass const reference to primitive", function() {
					            assert.equal(3, cm.const_ref_adder(1, 2));
					        });

					        test("get instance pointer as value", function() {
					            var v = cm.emval_test_instance_pointer();
					            assert.instanceof(v, cm.DummyForPointer);
					        });

					        test("cast value to instance pointer using as<T*>", function() {
					            var v = cm.emval_test_instance_pointer();
					            var p_value = cm.emval_test_value_from_instance_pointer(v);
					            assert.equal(42, p_value);
					        });

					        test("passthrough", function() {
					            var a = {foo: 'bar'};
					            var b = cm.emval_test_passthrough(a);
					            a.bar = 'baz';
					            assert.equal('baz', b.bar);

					            assert.equal(0, cm.count_emval_handles());
					        });

					        test("void return converts to undefined", function() {
					            assert.equal(undefined, cm.emval_test_return_void());
					        });

					        test("booleans can be marshalled", function() {
					            assert.equal(false, cm.emval_test_not(true));
					            assert.equal(true, cm.emval_test_not(false));
					        });

					        test("val.is_undefined() is functional",function() {
					            assert.equal(true, cm.emval_test_is_undefined(undefined));
					            assert.equal(false, cm.emval_test_is_undefined(true));
					            assert.equal(false, cm.emval_test_is_undefined(false));
					            assert.equal(false, cm.emval_test_is_undefined(null));
					            assert.equal(false, cm.emval_test_is_undefined({}));
					        });

					        test("val.is_null() is functional",function() {
					            assert.equal(true, cm.emval_test_is_null(null));
					            assert.equal(false, cm.emval_test_is_null(true));
					            assert.equal(false, cm.emval_test_is_null(false));
					            assert.equal(false, cm.emval_test_is_null(undefined));
					            assert.equal(false, cm.emval_test_is_null({}));
					        });

					        test("val.is_true() is functional",function() {
					            assert.equal(true, cm.emval_test_is_true(true));
					            assert.equal(false, cm.emval_test_is_true(false));
					            assert.equal(false, cm.emval_test_is_true(null));
					            assert.equal(false, cm.emval_test_is_true(undefined));
					            assert.equal(false, cm.emval_test_is_true({}));
					        });

					        test("val.is_false() is functional",function() {
					            assert.equal(true, cm.emval_test_is_false(false));
					            assert.equal(false, cm.emval_test_is_false(true));
					            assert.equal(false, cm.emval_test_is_false(null));
					            assert.equal(false, cm.emval_test_is_false(undefined));
					            assert.equal(false, cm.emval_test_is_false({}));
					        });

					        test("val.equals() is functional",function() {
					            var values = [undefined, null, true, false, {}];

					            for(var i=0;i<values.length;++i){
					                var first = values[i];
					                for(var j=i;j<values.length;++j)
					                {
					                    var second = values[j];
					assert.equal((first == second), cm.emval_test_equals(first, second));
				}
			}
			});

			test("val.strictlyEquals() is functional", function() {
			var values = [undefined, null, true, false, {}];

			for(var i=0;i<values.length;++i){
			var first = values[i];
			for(var j=i;j<values.length;++j)
			{
			var second = values[j];
			assert.equal(first===second, cm.emval_test_strictly_equals(first, second));
			}
			}
			});

			test("can pass booleans as integers", function() {
			assert.equal(1, cm.emval_test_as_unsigned(true));
			assert.equal(0, cm.emval_test_as_unsigned(false));
			});

			test("can pass booleans as floats", function() {
			assert.equal(2, cm.const_ref_adder(true, true));
			});

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
		/*
		   test("std::vector returns as an native object", function() {
		       var vec = cm.emval_test_return_vector();

		       assert.equal(3, vec.size());
		       assert.equal(10, vec.get(0));
		       assert.equal(20, vec.get(1));
		       assert.equal(30, vec.get(2));
		       vec.delete();
		   });

		   test("out of bounds std::vector access returns undefined", function() {
		       var vec = cm.emval_test_return_vector();

		       assert.equal(undefined, vec.get(4));
		       // only test a negative index without assertions.
		       if (!cm.getCompilerSetting('ASSERTIONS')) {
		           assert.equal(undefined, vec.get(-1));
		       }
		       vec.delete();
		   });

		   if (cm.getCompilerSetting('ASSERTIONS')) {
		       test("out of type range array index throws with assertions", function() {
		           var vec = cm.emval_test_return_vector();

		           assert.throws(TypeError, function() { vec.get(-1); });

		           vec.delete();
		       });
		   }

		   test("std::vector<std::shared_ptr<>> can be passed back", function() {
		       var vec = cm.emval_test_return_shared_ptr_vector();

		       assert.equal(2, vec.size());
		       var str0 = vec.get(0);
		       var str1 = vec.get(1);

		       assert.equal('string #1', str0.get());
		       assert.equal('string #2', str1.get());
		       str0.delete();
		       str1.delete();

		       vec.delete();
		   });

		   test("objects can be pushed back", function() {
		       var vectorHolder = new cm.VectorHolder();
		       var vec = vectorHolder.get();
		       assert.equal(2, vec.size());

		       var str = new cm.StringHolder('abc');
		       vec.push_back(str);
		       str.delete();
		       assert.equal(3, vec.size());
		       var str = vec.get(2);
		       assert.equal('abc', str.get());

		       str.delete();
		       vec.delete();
		       vectorHolder.delete();
		   });

		   test("can get elements with array operator", function(){
		       var vec = cm.emval_test_return_vector();
		       assert.equal(10, vec.get(0));
		       vec.delete();
		   });

		   test("can set elements with array operator", function() {
		       var vec = cm.emval_test_return_vector();
		       assert.equal(10, vec.get(0));
		       vec.set(2, 60);
		       assert.equal(60, vec.get(2));
		       vec.delete();
		   });

		   test("can set and get objects", function() {
		       var vec = cm.emval_test_return_shared_ptr_vector();
		       var str = vec.get(0);
		       assert.equal('string #1', str.get());
		       str.delete();
		       vec.delete();
		   });

		   test("resize appends the given value", function() {
		       var vec = cm.emval_test_return_vector();

		       vec.resize(5, 42);
		       assert.equal(5, vec.size());
		       assert.equal(10, vec.get(0));
		       assert.equal(20, vec.get(1));
		       assert.equal(30, vec.get(2));
		       assert.equal(42, vec.get(3));
		       assert.equal(42, vec.get(4));
		       vec.delete();
		   });

		   test("resize preserves content when shrinking", function() {
		       var vec = cm.emval_test_return_vector();

		       vec.resize(2, 42);
		       assert.equal(2, vec.size());
		       assert.equal(10, vec.get(0));
		       assert.equal(20, vec.get(1));
		       vec.delete();
		   });
		*/
	})

	When("vector", func() {
		/*
		   test("std::vector returns as an native object", function() {
		       var vec = cm.emval_test_return_vector();

		       assert.equal(3, vec.size());
		       assert.equal(10, vec.get(0));
		       assert.equal(20, vec.get(1));
		       assert.equal(30, vec.get(2));
		       vec.delete();
		   });

		   test("out of bounds std::vector access returns undefined", function() {
		       var vec = cm.emval_test_return_vector();

		       assert.equal(undefined, vec.get(4));
		       // only test a negative index without assertions.
		       if (!cm.getCompilerSetting('ASSERTIONS')) {
		           assert.equal(undefined, vec.get(-1));
		       }
		       vec.delete();
		   });

		   if (cm.getCompilerSetting('ASSERTIONS')) {
		       test("out of type range array index throws with assertions", function() {
		           var vec = cm.emval_test_return_vector();

		           assert.throws(TypeError, function() { vec.get(-1); });

		           vec.delete();
		       });
		   }

		   test("std::vector<std::shared_ptr<>> can be passed back", function() {
		       var vec = cm.emval_test_return_shared_ptr_vector();

		       assert.equal(2, vec.size());
		       var str0 = vec.get(0);
		       var str1 = vec.get(1);

		       assert.equal('string #1', str0.get());
		       assert.equal('string #2', str1.get());
		       str0.delete();
		       str1.delete();

		       vec.delete();
		   });

		   test("objects can be pushed back", function() {
		       var vectorHolder = new cm.VectorHolder();
		       var vec = vectorHolder.get();
		       assert.equal(2, vec.size());

		       var str = new cm.StringHolder('abc');
		       vec.push_back(str);
		       str.delete();
		       assert.equal(3, vec.size());
		       var str = vec.get(2);
		       assert.equal('abc', str.get());

		       str.delete();
		       vec.delete();
		       vectorHolder.delete();
		   });

		   test("can get elements with array operator", function(){
		       var vec = cm.emval_test_return_vector();
		       assert.equal(10, vec.get(0));
		       vec.delete();
		   });

		   test("can set elements with array operator", function() {
		       var vec = cm.emval_test_return_vector();
		       assert.equal(10, vec.get(0));
		       vec.set(2, 60);
		       assert.equal(60, vec.get(2));
		       vec.delete();
		   });

		   test("can set and get objects", function() {
		       var vec = cm.emval_test_return_shared_ptr_vector();
		       var str = vec.get(0);
		       assert.equal('string #1', str.get());
		       str.delete();
		       vec.delete();
		   });

		   test("resize appends the given value", function() {
		       var vec = cm.emval_test_return_vector();

		       vec.resize(5, 42);
		       assert.equal(5, vec.size());
		       assert.equal(10, vec.get(0));
		       assert.equal(20, vec.get(1));
		       assert.equal(30, vec.get(2));
		       assert.equal(42, vec.get(3));
		       assert.equal(42, vec.get(4));
		       vec.delete();
		   });

		   test("resize preserves content when shrinking", function() {
		       var vec = cm.emval_test_return_vector();

		       vec.resize(2, 42);
		       assert.equal(2, vec.size());
		       assert.equal(10, vec.get(0));
		       assert.equal(20, vec.get(1));
		       vec.delete();
		   });

		*/
	})

	When("map", func() {
		/*
		   test("std::map returns as native object", function() {
		       var map = cm.embind_test_get_string_int_map();

		       assert.equal(2, map.size());
		       assert.equal(1, map.get("one"));
		       assert.equal(2, map.get("two"));

		       map.delete();
		   });

		   test("std::map can get keys", function() {
		       var map = cm.embind_test_get_string_int_map();

		       var keys = map.keys();
		       assert.equal(map.size(), keys.size());
		       assert.equal("one", keys.get(0));
		       assert.equal("two", keys.get(1));
		       keys.delete();

		       map.delete();
		   });

		   test("std::map can set keys and values", function() {
		       var map = cm.embind_test_get_string_int_map();

		       assert.equal(2, map.size());

		       map.set("three", 3);

		       assert.equal(3, map.size());
		       assert.equal(3, map.get("three"));

		       map.set("three", 4);

		       assert.equal(3, map.size());
		       assert.equal(4, map.get("three"));

		       map.delete();
		   });

		*/
	})

	When("functors", func() {
		/*
		   test("can get and call function ptrs", function() {
		       var ptr = cm.emval_test_get_function_ptr();
		       assert.equal("foobar", ptr.opcall("foobar"));
		       ptr.delete();
		   });

		   test("can pass functor to C++", function() {
		       var ptr = cm.emval_test_get_function_ptr();
		       assert.equal("asdf", cm.emval_test_take_and_call_functor(ptr));
		       ptr.delete();
		   });

		   test("can clone handles", function() {
		       var a = cm.emval_test_get_function_ptr();
		       var b = a.clone();
		       a.delete();

		       assert.throws(cm.BindingError, function() {
		           a.delete();
		       });
		       b.delete();
		   });

		*/
	})

	When("classes", func() {
		/*
			        test("class instance", function() {
			            var a = {foo: 'bar'};
			            assert.equal(0, cm.count_emval_handles());
			            var c = new cm.ValHolder(a);
			            assert.equal(1, cm.count_emval_handles());
			            assert.equal('bar', c.getVal().foo);
			            assert.equal(1, cm.count_emval_handles());

			            c.setVal('1234');
			            assert.equal('1234', c.getVal());

			            c.delete();
			            assert.equal(0, cm.count_emval_handles());
			        });

			        test("class properties can be methods", function() {
			            var a = {};
			            var b = {foo: 'foo'};
			            var c = new cm.ValHolder(a);
			            assert.equal(a, c.val);
			            c.val = b;
			            assert.equal(b, c.val);
			            c.delete();
			        });

			        test("class properties can be std::function objects", function() {
			            var a = {};
			            var b = {foo: 'foo'};
			            var c = new cm.ValHolder(a);
			            assert.equal(a, c.function_val);
			            c.function_val = b;
			            assert.equal(b, c.function_val);
			            c.delete();
			        });

			        test("class properties can be read-only std::function objects", function() {
			            var a = {};
			            var h = new cm.ValHolder(a);
			            assert.equal(a, h.readonly_function_val);
			            var e = assert.throws(cm.BindingError, function() {
			                h.readonly_function_val = 10;
			            });
			            assert.equal('ValHolder.readonly_function_val is a read-only property', e.message);
			            h.delete();
			        });

			        test("class properties can be function objects (functor)", function() {
			            var a = {};
			            var b = {foo: 'foo'};
			            var c = new cm.ValHolder(a);
			            assert.equal(a, c.functor_val);
			            c.function_val = b;
			            assert.equal(b, c.functor_val);
			            c.delete();
			        });

			        test("class properties can be read-only function objects (functor)", function() {
			            var a = {};
			            var h = new cm.ValHolder(a);
			            assert.equal(a, h.readonly_functor_val);
			            var e = assert.throws(cm.BindingError, function() {
			                h.readonly_functor_val = 10;
			            });
			            assert.equal('ValHolder.readonly_functor_val is a read-only property', e.message);
			            h.delete();
			        });

			        test("class properties can be read-only", function() {
			            var a = {};
			            var h = new cm.ValHolder(a);
			            assert.equal(a, h.val_readonly);
			            var e = assert.throws(cm.BindingError, function() {
			                h.val_readonly = 10;
			            });
			            assert.equal('ValHolder.val_readonly is a read-only property', e.message);
			            h.delete();
			        });

			        test("read-only member field", function() {
			            var a = new cm.HasReadOnlyProperty(10);
			            assert.equal(10, a.i);
			            var e = assert.throws(cm.BindingError, function() {
			                a.i = 20;
			            });
			            assert.equal('HasReadOnlyProperty.i is a read-only property', e.message);
			            a.delete();
			        });

			        test("class instance $$ property is non-enumerable", function() {
			            var c = new cm.ValHolder(undefined);
			            assert.deepEqual([], Object.keys(c));
			            var d = c.clone();
			            c.delete();

			            assert.deepEqual([], Object.keys(d));
			            d.delete();
			        });

			        test("class methods", function() {
			            assert.equal(10, cm.ValHolder.some_class_method(10));

			            var b = cm.ValHolder.makeValHolder("foo");
			            assert.equal("foo", b.getVal());
			            b.delete();
			        });

			        test("function objects as class constructors", function() {
			            var a = new cm.ConstructFromStdFunction("foo", 10);
			            assert.equal("foo", a.getVal());
			            assert.equal(10, a.getA());

			            var b = new cm.ConstructFromFunctionObject("bar", 12);
			            assert.equal("bar", b.getVal());
			            assert.equal(12, b.getA());

			            a.delete();
			            b.delete();
			        });

			        test("function objects as class methods", function() {
			            var b = cm.ValHolder.makeValHolder("foo");

			            // get & set via std::function
			            assert.equal("foo", b.getValFunction());
			            b.setValFunction("bar");

			            // get & set via 'callable'
			            assert.equal("bar", b.getValFunctor());
			            b.setValFunctor("baz");

			            assert.equal("baz", b.getValFunction());

			            b.delete();
			        });

			        test("can't call methods on deleted class instances", function() {
			            var c = new cm.ValHolder(undefined);
			            c.delete();
			            assert.throws(cm.BindingError, function() {
			                c.getVal();
			            });
			            assert.throws(cm.BindingError, function() {
			                c.delete();
			            });
			        });

			        test("calling constructor without new raises BindingError", function() {
			            var e = assert.throws(cm.BindingError, function() {
			                cm.ValHolder(undefined);
			            });
			            assert.equal("Use 'new' to construct ValHolder", e.message);
			        });

			        test("can return class instances by value", function() {
			            var c = cm.emval_test_return_ValHolder();
			            assert.deepEqual({}, c.getVal());
			            c.delete();
			        });

			        test("can pass class instances to functions by reference", function() {
			            var a = {a:1};
			            var c = new cm.ValHolder(a);
			            cm.emval_test_set_ValHolder_to_empty_object(c);
			            assert.deepEqual({}, c.getVal());
			            c.delete();
			        });

			        test("can pass smart pointer by reference", function() {
			            var base = cm.embind_test_return_smart_base_ptr();
			            var name = cm.embind_test_get_class_name_via_reference_to_smart_base_ptr(base);
			            assert.equal("Base", name);
			            base.delete();
			        });

			        test("can pass smart pointer by value", function() {
			            var base = cm.embind_test_return_smart_base_ptr();
			            var name = cm.embind_test_get_class_name_via_smart_base_ptr(base);
			            assert.equal("Base", name);
			            base.delete();
			        });

			        // todo: fix this
			        // This test does not work because we make no provision for argument values
			        // having been changed after returning from a C++ routine invocation. In
			        // this specific case, the original pointee of the smart pointer was
			        // freed and replaced by a new one, but the ptr in our local handle
			        // was never updated after returning from the call.
			        test("can modify smart pointers passed by reference", function() {
			//            var base = cm.embind_test_return_smart_base_ptr();
			//            cm.embind_modify_smart_pointer_passed_by_reference(base);
			//            assert.equal("Changed", base.getClassName());
			//            base.delete();
			        });

			        test("can not modify smart pointers passed by value", function() {
			            var base = cm.embind_test_return_smart_base_ptr();
			            cm.embind_attempt_to_modify_smart_pointer_when_passed_by_value(base);
			            assert.equal("Base", base.getClassName());
			            base.delete();
			        });

			        test("const return value", function() {
			            var c = new cm.ValHolder("foo");
			            assert.equal("foo", c.getConstVal());
			            c.delete();
			        });

			        test("return object by const ref", function() {
			            var c = new cm.ValHolder("foo");
			            assert.equal("foo", c.getValConstRef());
			            c.delete();
			        });

			        test("instanceof", function() {
			            var c = new cm.ValHolder("foo");
			            assert.instanceof(c, cm.ValHolder);
			            c.delete();
			        });

			        test("can access struct fields", function() {
			            var c = new cm.CustomStruct();
			            assert.equal(10, c.field);
			            assert.equal(10, c.getField());
			            c.delete();
			        });

			        test("can set struct fields", function() {
			            var c = new cm.CustomStruct();
			            c.field = 15;
			            assert.equal(15, c.field);
			            c.delete();
			        });

			        test("assignment returns value", function() {
			            var c = new cm.CustomStruct();
			            assert.equal(15, c.field = 15);
			            c.delete();
			        });

			        if (cm.getCompilerSetting('ASSERTIONS')) {
			            test("assigning string or object to integer raises TypeError with assertions", function() {
			                var c = new cm.CustomStruct();
			                var e = assert.throws(TypeError, function() {
			                    c.field = "hi";
			                });
			                assert.equal('Cannot convert "hi" to int', e.message);

			                var e = assert.throws(TypeError, function() {
			                    c.field = {foo:'bar'};
			                });
			                assert.equal('Cannot convert "[object Object]" to int', e.message);

			                c.delete();
			            });
			        } else {
			            test("assigning string or object to integer is converted to 0", function() {
			                var c = new cm.CustomStruct();

			                c.field = "hi";
			                assert.equal(0, c.field);
			                c.field = {foo:'bar'};
			                assert.equal(0, c.field);

			                c.delete();
			            });
			        }

			        test("can return tuples by value", function() {
			            var c = cm.emval_test_return_TupleVector();
			            assert.deepEqual([1, 2, 3, 4], c);
			        });

			        test("tuples can contain tuples", function() {
			            var c = cm.emval_test_return_TupleVectorTuple();
			            assert.deepEqual([[1, 2, 3, 4]], c);
			        });

			        test("can pass tuples by value", function() {
			            var c = cm.emval_test_take_and_return_TupleVector([4, 5, 6, 7]);
			            assert.deepEqual([4, 5, 6, 7], c);
			        });

			        test("can return structs by value", function() {
			            var c = cm.emval_test_return_StructVector();
			            assert.deepEqual({x: 1, y: 2, z: 3, w: 4}, c);
			        });

			        test("can pass structs by value", function() {
			            var c = cm.emval_test_take_and_return_StructVector({x: 4, y: 5, z: 6, w: 7});
			            assert.deepEqual({x: 4, y: 5, z: 6, w: 7}, c);
			        });

			        test("can pass and return tuples in structs", function() {
			            var d = cm.emval_test_take_and_return_TupleInStruct({field: [1, 2, 3, 4]});
			            assert.deepEqual({field: [1, 2, 3, 4]}, d);
			        });

			        test("can pass and return arrays in structs", function() {
			            var d = cm.emval_test_take_and_return_ArrayInStruct({
			              field1: [1, 2],
			              field2: [
			                { x: 1, y: 2 },
			                { x: 3, y: 4 }
			              ]
			            });
			            assert.deepEqual({
			              field1: [1, 2],
			              field2: [
			                { x: 1, y: 2 },
			                { x: 3, y: 4 }
			              ]
			            }, d);
			        });

			        test("can clone handles", function() {
			            var a = new cm.ValHolder({});
			            assert.equal(1, cm.count_emval_handles());
			            var b = a.clone();
			            a.delete();

			            assert.equal(1, cm.count_emval_handles());

			            assert.throws(cm.BindingError, function() {
			                a.delete();
			            });
			            b.delete();

			            assert.equal(0, cm.count_emval_handles());
			        });

			        test("A shared pointer set/get point to the same underlying pointer", function() {
			            var a = new cm.SharedPtrHolder();
			            var b = a.get();

			            a.set(b);
			            var c = a.get();

			            assert.true(b.isAliasOf(c));
			            b.delete();
			            c.delete();
			            a.delete();
			        });

			        test("can return shared ptrs from instance methods", function() {
			            var a = new cm.SharedPtrHolder();

			            // returns the shared_ptr.
			            var b = a.get();

			            assert.equal("a string", b.get());
			            b.delete();
			            a.delete();
			        });

			        test("smart ptrs clone correctly", function() {
			            assert.equal(0, cm.count_emval_handles());

			            var a = cm.emval_test_return_shared_ptr();

			            var b = a.clone();
			            a.delete();

			            assert.equal(1, cm.count_emval_handles());

			            assert.throws(cm.BindingError, function() {
			                a.delete();
			            });
			            b.delete();

			            assert.equal(0, cm.count_emval_handles());
			        });

			        test("can't clone if already deleted", function() {
			            var a = new cm.ValHolder({});
			            a.delete();
			            assert.throws(cm.BindingError, function() {
			                a.clone();
			            });
			        });

			        test("virtual calls work correctly", function() {
			            var derived = cm.embind_test_return_raw_polymorphic_derived_ptr_as_base();
			            assert.equal("PolyDerived", derived.virtualGetClassName());
			            derived.delete();
			        });

			        test("virtual calls work correctly on smart ptrs", function() {
			            var derived = cm.embind_test_return_smart_polymorphic_derived_ptr_as_base();
			            assert.equal("PolyDerived", derived.virtualGetClassName());
			            derived.delete();
			        });

			        test("Empty smart ptr is null", function() {
			            var a = cm.emval_test_return_empty_shared_ptr();
			            assert.equal(null, a);
			        });

			        test("string cannot be given as smart pointer argument", function() {
			            assert.throws(cm.BindingError, function() {
			                cm.emval_test_is_shared_ptr_null("hello world");
			            });
			        });

			        test("number cannot be given as smart pointer argument", function() {
			            assert.throws(cm.BindingError, function() {
			                cm.emval_test_is_shared_ptr_null(105);
			            });
			        });

			        test("raw pointer cannot be given as smart pointer argument", function() {
			            var p = new cm.ValHolder({});
			            assert.throws(cm.BindingError, function() { cm.emval_test_is_shared_ptr_null(p); });
			            p.delete();
			        });

			        test("null is passed as empty smart pointer", function() {
			            assert.true(cm.emval_test_is_shared_ptr_null(null));
			        });

			        test("Deleting already deleted smart ptrs fails", function() {
			            var a = cm.emval_test_return_shared_ptr();
			            a.delete();
			            assert.throws(cm.BindingError, function() {
			                a.delete();
			            });
			        });

			        test("returned unique_ptr does not call destructor", function() {
			            var logged = "";
			            var c = new cm.emval_test_return_unique_ptr_lifetime(function (s) { logged += s; });
			            assert.equal("(constructor)", logged);
			            c.delete();
			        });

			        test("returned unique_ptr calls destructor on delete", function() {
			            var logged = "";
			            var c = new cm.emval_test_return_unique_ptr_lifetime(function (s) { logged += s; });
			            logged = "";
			            c.delete();
			            assert.equal("(destructor)", logged);
			        });

			        test("StringHolder", function() {
			            var a = new cm.StringHolder("foobar");
			            assert.equal("foobar", a.get());

			            a.set("barfoo");
			            assert.equal("barfoo", a.get());

			            assert.equal("barfoo", a.get_const_ref());

			            a.delete();
			        });

			        test("can call methods on unique ptr", function() {
			            var result = cm.emval_test_return_unique_ptr();

			            result.setVal('1234');
			            assert.equal('1234', result.getVal());
			            result.delete();
			        });

			        test("can call methods on shared ptr", function(){
			            var result = cm.emval_test_return_shared_ptr();
			            result.setVal('1234');

			            assert.equal('1234', result.getVal());
			            result.delete();
			        });

			        test("Non functors throw exception", function() {
			            var a = {foo: 'bar'};
			            var c = new cm.ValHolder(a);
			            assert.throws(TypeError, function() {
			                c();
			            });
			            c.delete();
			        });

			        test("non-member methods", function() {
			            var a = {foo: 'bar'};
			            var c = new cm.ValHolder(a);
			            c.setEmpty(); // non-member method
			            assert.deepEqual({}, c.getValNonMember());
			            c.delete();
			        });

			        test("instantiating class without constructor gives error", function() {
			            var e = assert.throws(cm.BindingError, function() {
			                cm.AbstractClass();
			            });
			            assert.equal("Use 'new' to construct AbstractClass", e.message);

			            var e = assert.throws(cm.BindingError, function() {
			                new cm.AbstractClass();
			            });
			            assert.equal("AbstractClass has no accessible constructor", e.message);
			        });

			        test("can construct class with external constructor", function() {
			            var e = new cm.HasExternalConstructor("foo");
			            assert.instanceof(e, cm.HasExternalConstructor);
			            assert.equal("foo", e.getString());
			            e.delete();
			        });
		*/
	})

	When("const", func() {
		/*
		   test("calling non-const method with const handle is error", function() {
		       var vh = cm.ValHolder.makeConst({});
		       var e = assert.throws(cm.BindingError, function() {
		           vh.setVal({});
		       });
		       assert.equal('Cannot convert argument of type ValHolder const* to parameter type ValHolder*', e.message);
		       vh.delete();
		   });

		   test("passing const pointer to non-const pointer is error", function() {
		       var vh = new cm.ValHolder.makeConst({});
		       var e = assert.throws(cm.BindingError, function() {
		           cm.ValHolder.set_via_raw_pointer(vh, {});
		       });
		       assert.equal('Cannot convert argument of type ValHolder const* to parameter type ValHolder*', e.message);
		       vh.delete();
		   });

		*/
	})

	When("smart pointers", func() {
		/*
		   test("constructor can return smart pointer", function() {
		       var e = new cm.HeldBySmartPtr(10, "foo");
		       assert.instanceof(e, cm.HeldBySmartPtr);
		       assert.equal(10, e.i);
		       assert.equal("foo", e.s);
		       var f = cm.takesHeldBySmartPtr(e);
		       f.delete();
		       e.delete();
		   });

		   test("cannot pass incorrect smart pointer type", function() {
		       var e = cm.emval_test_return_shared_ptr();
		       assert.throws(cm.BindingError, function() {
		           cm.takesHeldBySmartPtr(e);
		       });
		       e.delete();
		   });

		   test("smart pointer object has no object keys", function() {
		       var e = new cm.HeldBySmartPtr(10, "foo");
		       assert.deepEqual([], Object.keys(e));

		       var f = e.clone();
		       e.delete();

		       assert.deepEqual([], Object.keys(f));
		       f.delete();
		   });

		   test("smart pointer object has correct constructor name", function() {
		       var e = new cm.HeldBySmartPtr(10, "foo");
		       assert.equal("HeldBySmartPtr", e.constructor.name);
		       e.delete();
		   });

		   test("constructor can return smart pointer", function() {
		       var e = new cm.HeldBySmartPtr(10, "foo");
		       assert.instanceof(e, cm.HeldBySmartPtr);
		       assert.equal(10, e.i);
		       assert.equal("foo", e.s);
		       var f = cm.takesHeldBySmartPtr(e);
		       assert.instanceof(f, cm.HeldBySmartPtr);
		       f.delete();
		       e.delete();
		   });

		   test("custom smart pointer", function() {
		       var e = new cm.HeldByCustomSmartPtr(20, "bar");
		       assert.instanceof(e, cm.HeldByCustomSmartPtr);
		       assert.equal(20, e.i);
		       assert.equal("bar", e.s);
		       e.delete();
		   });

		   test("custom smart pointer passed through wiretype", function() {
		       var e = new cm.HeldByCustomSmartPtr(20, "bar");
		       var f = cm.passThroughCustomSmartPtr(e);
		       e.delete();

		       assert.instanceof(f, cm.HeldByCustomSmartPtr);
		       assert.equal(20, f.i);
		       assert.equal("bar", f.s);

		       f.delete();
		   });

		   test("cannot give null to by-value argument", function() {
		       var e = assert.throws(cm.BindingError, function() {
		           cm.takesHeldBySmartPtr(null);
		       });
		       assert.equal('null is not a valid HeldBySmartPtr', e.message);
		   });

		   test("raw pointer can take and give null", function() {
		       assert.equal(null, cm.passThroughRawPtr(null));
		   });

		   test("custom smart pointer can take and give null", function() {
		       assert.equal(null, cm.passThroughCustomSmartPtr(null));
		   });

		   test("cannot pass shared_ptr to CustomSmartPtr", function() {
		       var o = cm.HeldByCustomSmartPtr.createSharedPtr(10, "foo");
		       var e = assert.throws(cm.BindingError, function() {
		           cm.passThroughCustomSmartPtr(o);
		       });
		       assert.equal('Cannot convert argument of type shared_ptr<HeldByCustomSmartPtr> to parameter type CustomSmartPtr<HeldByCustomSmartPtr>', e.message);
		       o.delete();
		   });

		   test("custom smart pointers can be passed to shared_ptr parameter", function() {
		       var e = cm.HeldBySmartPtr.newCustomPtr(10, "abc");
		       assert.equal(10, e.i);
		       assert.equal("abc", e.s);

		       cm.takesHeldBySmartPtrSharedPtr(e).delete();
		       e.delete();
		   });

		   test("can call non-member functions as methods", function() {
		       var e = new cm.HeldBySmartPtr(20, "bar");
		       var f = e.returnThis();
		       e.delete();
		       assert.equal(20, f.i);
		       assert.equal("bar", f.s);
		       f.delete();
		   });
		*/
	})

	When("enumerations", func() {
		/*
		   test("can compare enumeration values", function() {
		       assert.equal(cm.Enum.ONE, cm.Enum.ONE);
		       assert.notEqual(cm.Enum.ONE, cm.Enum.TWO);
		   });

		   if (typeof INVOKED_FROM_EMSCRIPTEN_TEST_RUNNER === "undefined") { // TODO: Enable this to work in Emscripten runner as well!
		       test("repr includes enum value", function() {
		           assert.equal('<#Enum_ONE {}>', IMVU.repr(cm.Enum.ONE));
		           assert.equal('<#Enum_TWO {}>', IMVU.repr(cm.Enum.TWO));
		       });
		   }

		   test("instanceof", function() {
		       assert.instanceof(cm.Enum.ONE, cm.Enum);
		   });

		   test("can pass and return enumeration values to functions", function() {
		       assert.equal(cm.Enum.TWO, cm.emval_test_take_and_return_Enum(cm.Enum.TWO));
		   });

		*/
	})

	When("C++11 enum class", func() {
		/*
		       test("can compare enumeration values", function() {
		           assert.equal(cm.EnumClass.ONE, cm.EnumClass.ONE);
		           assert.notEqual(cm.EnumClass.ONE, cm.EnumClass.TWO);
		       });

		       if (typeof INVOKED_FROM_EMSCRIPTEN_TEST_RUNNER === "undefined") { // TODO: Enable this to work in Emscripten runner as well!
		           test("repr includes enum value", function() {
		               assert.equal('<#EnumClass_ONE {}>', IMVU.repr(cm.EnumClass.ONE));
		               assert.equal('<#EnumClass_TWO {}>', IMVU.repr(cm.EnumClass.TWO));
		           });
		       }

		       test("instanceof", function() {
		           assert.instanceof(cm.EnumClass.ONE, cm.EnumClass);
		       });

		       test("can pass and return enumeration values to functions", function() {
		           assert.equal(cm.EnumClass.TWO, cm.emval_test_take_and_return_EnumClass(cm.EnumClass.TWO));
		       });
		   });
		*/
	})

	When("emval call tests", func() {
		/*
		   test("can call functions from C++", function() {
		       var called = false;
		       cm.emval_test_call_function(function(i, f, tv, sv) {
		           called = true;
		           assert.equal(10, i);
		           assert.equal(1.5, f);
		           assert.deepEqual([1.25, 2.5, 3.75, 4], tv);
		           assert.deepEqual({x: 1.25, y: 2.5, z: 3.75, w:4}, sv);
		       }, 10, 1.5, [1.25, 2.5, 3.75, 4], {x: 1.25, y: 2.5, z: 3.75, w:4});
		       assert.true(called);
		   });
		*/
	})

	When("extending built-in classes", func() {
		/*
		   // cm.ValHolder.prototype.patched = 10; // this sets instanceCounts.patched inside of Deletable module !?!

		   test("can access patched value on new instances", function() {
		       cm.ValHolder.prototype.patched = 10;
		       var c = new cm.ValHolder(undefined);
		       assert.equal(10, c.patched);
		       c.delete();
		       cm.ValHolder.prototype.patched = undefined;
		   });

		   test("can access patched value on returned instances", function() {
		       cm.ValHolder.prototype.patched = 10;
		       var c = cm.emval_test_return_ValHolder();
		       assert.equal(10, c.patched);
		       c.delete();
		       cm.ValHolder.prototype.patched = undefined;
		   });

		*/
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

		   test("emscripten::val method arguments don't leak", function() {
		       var parent = cm.AbstractClass;
		       var got;
		       var C = parent.extend("C", {
		           abstractMethod: function() {
		           },
		           passVal: function(g) {
		               got = g;
		           }
		       });
		       var impl = new C;
		       var v = {};
		       cm.passVal(impl, v);
		       impl.delete();

		       assert.equal(v, got);
		   });
		*/
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

	When("function names", func() {
		// @todo: do these make sense in Go?
		/*
		   assert.equal('ValHolder', cm.ValHolder.name);

		   if (!cm.getCompilerSetting('DYNAMIC_EXECUTION')) {
		     assert.equal('', cm.ValHolder.prototype.setVal.name);
		     assert.equal('', cm.ValHolder.makeConst.name);
		   } else {
		     assert.equal('ValHolder$setVal', cm.ValHolder.prototype.setVal.name);
		     assert.equal('ValHolder$makeConst', cm.ValHolder.makeConst.name);
		   }
		*/
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

		It("before and after memory growth", func() {
			// @todo: implement EmvalNewArray.
			// @todo: implement some globals that we can also have on our side (like Uint8Array).
			//array, err := generated.Construct_with_arguments_before_and_after_memory_growth(engine, ctx)
			//Expect(err).To(BeNil())
			//Expect(array.([]uint8)[0]).To(HaveLen(5))
			//Expect(array.([]uint8)[0]).To(HaveLen(len(array.([]uint8)[1])))
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

		// @todo: implement me
		// We don't support this in Go. Needs CreateInheritingConstructor.
		It("can extend from intrusive pointer class and still preserve reference in JavaScript", func() {
			//type newStructTypeToExtend struct {
			//	embind.ClassBase
			//}
			//C, err := generated.ClassIntrusiveClassStaticExtend(engine, ctx, "C2", &newStructTypeToExtend{})
			//Expect(err).To(BeNil())
			//log.Println(C)
			//log.Println(C.(func(context.Context, ...any) (any, error))(ctx))

			//var instance = new C;
			//var holder = new cm.IntrusiveClassHolder;
			//holder.set(instance);
			//instance.delete();

			//var back = holder.get();
			//assert.equal(back, instance);
			//holder.delete();
			//back.delete();
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
