#include <emscripten/bind.h>

using namespace emscripten;

enum OldStyle {
    OLD_STYLE_ONE,
    OLD_STYLE_TWO
};

enum class NewStyle {
    ONE,
    TWO
};

OldStyle enum_in_enum_out(NewStyle ns) {
    return OldStyle::OLD_STYLE_TWO;
}

EMSCRIPTEN_BINDINGS(enums) {
    enum_<OldStyle>("OldStyle")
        .value("ONE", OLD_STYLE_ONE)
        .value("TWO", OLD_STYLE_TWO)
        ;
    enum_<NewStyle>("NewStyle")
        .value("ONE", NewStyle::ONE)
        .value("TWO", NewStyle::TWO)
        ;
   function("enum_in_enum_out", &enum_in_enum_out);
}