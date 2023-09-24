#include <emscripten/bind.h>

using namespace emscripten;

void hello_world(std::string who) {
    printf("Hello world and %s!\n", who.c_str());
}

void function_overload() {
}

void function_overload_2(int x) {
}

EMSCRIPTEN_BINDINGS(my_module) {
    function("hello_world", &hello_world);

    function("function_overload", &function_overload);
    function("function_overload", &function_overload_2);
}