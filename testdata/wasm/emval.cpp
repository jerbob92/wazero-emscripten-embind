#include <emscripten/bind.h>
#include <emscripten/val.h>
#include <emscripten/version.h>
#include <stdio.h>
#include <math.h>

using namespace emscripten;

std::string doEmval() {
  std::string result = "";
  val AudioContext = val::global("AudioContext");
  if (!AudioContext.as<bool>()) {
    result += "No global AudioContext, trying webkitAudioContext\n";
    AudioContext = val::global("webkitAudioContext");
  }

  result += "Got an AudioContext\n";
  val context = AudioContext.new_();
  val oscillator = context.call<val>("createOscillator");

  result += "Configuring oscillator\n";
  oscillator.set("type", val("triangle"));
  oscillator["frequency"].set("value", val(261.63)); // Middle C

  result += "Playing\n";
  oscillator.call<void>("connect", context["destination"]);
  oscillator.call<void>("start", 0);

  result += "All done!\n";
  return result;
}

bool emval_instance_of(const val& v, const val& v2) {
    return v.instanceof(v2);
}

val emval_type_of(const val& v) {
    return v.typeOf();
}

bool emval_in(const val& v, const val& v2) {
    return v.in(v2);
}

void emval_throw(const val& v) {
    return v.throw_();
}

void emval_delete(const val& v) {
    v.delete_("test");
}

val emval_await(const val& v) {
    return v.await();
}

bool emval_is_number(const val& v) {
    return v.isNumber();
}

bool emval_is_string(const val& v) {
    return v.isString();
}

bool emval_is_array(const val& v) {
    return v.isArray();
}

bool emval_has_own_property(const val& v, const char* key) {
    return v.hasOwnProperty(key);
}

val emval_u16_string(const char16_t* s) {
    return val::u16string(s);
}

val emval_u8_string(const char* s) {
    return val::u8string(s);
}

val emval_array() {
    return val::array();
}

val emscripten_version() {
    std::vector<int> version_vec;
    version_vec.push_back(__EMSCRIPTEN_major__);
    version_vec.push_back(__EMSCRIPTEN_minor__);
    version_vec.push_back(__EMSCRIPTEN_tiny__);
    return val::array(version_vec);
}

#if __EMSCRIPTEN_major__ > 3 || (__EMSCRIPTEN_major__ == 3 && __EMSCRIPTEN_minor__ > 1) || (__EMSCRIPTEN_major__ == 3 && __EMSCRIPTEN_minor__ == 1 && __EMSCRIPTEN_tiny__ >= 47)
std::vector<int> emval_iterator() {
    std::vector<int> vec2;
    vec2.push_back(0);
    vec2.push_back(1);
    vec2.push_back(3);
    val::global().set("a", val::array(vec2));
    std::vector<int> vec2_from_iter;
    for (val&& v : val::global("a")) {
        vec2_from_iter.push_back(v.as<int>());
    }
    return vec2_from_iter;
}
#endif

EMSCRIPTEN_BINDINGS(emval) {
    function("doEmval", &doEmval);
    function("emval_instance_of", &emval_instance_of);
    function("emval_type_of", &emval_type_of);
    function("emval_in", &emval_in);
    function("emval_throw", &emval_throw);
    function("emval_delete", &emval_delete);
    function("emval_await", &emval_await);
    function("emval_is_number", &emval_is_number);
    function("emval_is_string", &emval_is_string);
    function("emval_is_array", &emval_is_array);
    function("emval_has_own_property", &emval_has_own_property, allow_raw_pointers());
    function("emval_u16_string", &emval_u16_string, allow_raw_pointers());
    function("emval_u8_string", &emval_u8_string, allow_raw_pointers());
    function("emval_array", &emval_array);
    function("emscripten_version", &emscripten_version);

    #if __EMSCRIPTEN_major__ > 3 || (__EMSCRIPTEN_major__ == 3 && __EMSCRIPTEN_minor__ > 1) || (__EMSCRIPTEN_major__ == 3 && __EMSCRIPTEN_minor__ == 1 && __EMSCRIPTEN_tiny__ >= 47)
    function("emval_iterator", &emval_iterator);
    #endif
}