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

struct Base {
    virtual void invoke(const std::string& str) {
        // default implementation
    }
};

struct BaseWrapper : public wrapper<Base> {
    EMSCRIPTEN_WRAPPER(BaseWrapper);
    void invoke(const std::string& str) {
        return call<void>("invoke", str);
    }
};

class Derived : public Base {};
Base* getDerivedInstance() {
    return new Derived;
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

      class_<Base>("Base")
        .allow_subclass<BaseWrapper>("BaseWrapper")
        .function("invoke", optional_override([](Base& self, const std::string& str) {
            return self.Base::invoke(str);
        }))
        ;

    class_<Derived, base<Base>>("Derived");
    function("getDerivedInstance", &getDerivedInstance, allow_raw_pointers());
}
