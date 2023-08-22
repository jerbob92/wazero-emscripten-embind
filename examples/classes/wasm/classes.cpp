#include <emscripten/bind.h>

using namespace emscripten;

class MyClass {
public:
  MyClass(int x, std::string y)
    : x(x)
    , y(y)
  {}

  void incrementX() {
    ++x;
  }

  int getX() const { return x; }
  void setX(int x_) { x = x_; }

  std::string getY() const { return y; }

  static std::string getStringFromInstance(const MyClass& instance) {
    return instance.y;
  }

protected:
  int x;
  std::string y;
};

void printMyClass(MyClass input) {
    printf("Have MyClass with x %d and y %s!\n", input.getX(), input.getY().c_str());
}

EMSCRIPTEN_BINDINGS(my_module) {
    class_<MyClass>("MyClass")
      .constructor<int, std::string>()
      .function("incrementX", &MyClass::incrementX)
      .property("x", &MyClass::getX, &MyClass::setX)
      .class_function("getStringFromInstance", &MyClass::getStringFromInstance)
      ;

    function("printMyClass", &printMyClass);
}