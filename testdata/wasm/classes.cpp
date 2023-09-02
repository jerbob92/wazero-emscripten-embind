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

  std::string yGetter() const { return y; }

  std::string getY(std::string in) const { return in + y; }

  static std::string getStringFromInstance(const MyClass& instance) {
    return instance.y;
  }

protected:
  int x;
  std::string y;
};

EMSCRIPTEN_BINDINGS(classes) {
    class_<MyClass>("MyClass")
      .constructor<int, std::string>()
      .constructor<int>()
      .function("incrementX", select_overload<void()>(&MyClass::incrementX))
      .function("incrementX", select_overload<void(int)>(&MyClass::incrementX))
      .function("getY", &MyClass::getY)
      .property("x", &MyClass::getX, &MyClass::setX)
      .property("y", &MyClass::yGetter)
      .class_function("getStringFromInstance", &MyClass::getStringFromInstance)
      ;
}
