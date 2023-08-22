#include <emscripten/bind.h>

using namespace emscripten;

void hello_world(std::string who) {
    printf("Hello world and %s!\n", who.c_str());
}

EMSCRIPTEN_BINDINGS(my_module) {
    function("hello_world", &hello_world);
}