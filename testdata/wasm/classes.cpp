#include <emscripten/bind.h>

using namespace emscripten;

class MyClass {
public:
  MyClass(int x, std::string y)
    : x(x)
    , y(y)
  {}

  MyClass(int x)
      : x(x)
    {}

  void incrementX() {
    ++x;
  }

  void incrementX(int multiplier) {
    x = x + (1 * multiplier);
  }

  int getX() const { return x; }
  void setX(int x_) { x = x_; }

  std::string getY() const { return y; }

  std::string combineY(std::string in) const { return in + y; }

  static std::string getStringFromInstance(const MyClass& instance) {
    return instance.y;
  }

protected:
  int x;
  std::string y;
};

MyClass* passThrough(MyClass* ptr) { return ptr; }

struct Interface {
    virtual void invoke(const std::string& str) = 0;
};

struct InterfaceWrapper : public wrapper<Interface> {
    EMSCRIPTEN_WRAPPER(InterfaceWrapper);
    void invoke(const std::string& str) {
        return call<void>("invoke", str);
    }
};

class C {};

struct BaseClass {
    virtual void invoke(const std::string& str) {
        // default implementation
    }
};

struct BaseClassWrapper : public wrapper<BaseClass> {
    EMSCRIPTEN_WRAPPER(BaseClassWrapper);
    void invoke(const std::string& str) {
        return call<void>("invoke", str);
    }
};

class DerivedClass : public BaseClass {};
BaseClass* getDerivedInstance() {
    return new DerivedClass;
}

EMSCRIPTEN_BINDINGS(classes) {
    class_<MyClass>("MyClass")
      .constructor<int, std::string>()
      .constructor<int>()
      .function("incrementX", select_overload<void()>(&MyClass::incrementX))
      .function("incrementX", select_overload<void(int)>(&MyClass::incrementX))
      .function("combineY", &MyClass::combineY)
      .property("x", &MyClass::getX, &MyClass::setX)
      .property("y", &MyClass::getY)
      .class_function("getStringFromInstance", &MyClass::getStringFromInstance)
      ;

      function("passThrough", &passThrough, allow_raw_pointers());

      class_<Interface>("Interface")
          .function("invoke", &Interface::invoke, pure_virtual())
          .allow_subclass<InterfaceWrapper>("InterfaceWrapper")
          ;

      class_<C>("C")
        .smart_ptr_constructor("C", &std::make_shared<C>)
        ;

      class_<BaseClass>("BaseClass")
        .allow_subclass<BaseClassWrapper>("BaseClassWrapper")
        .function("invoke", optional_override([](BaseClass& self, const std::string& str) {
            return self.BaseClass::invoke(str);
        }))
        ;

    class_<DerivedClass, base<BaseClass>>("DerivedClass");
    function("getDerivedClassInstance", &getDerivedInstance, allow_raw_pointers());
}
