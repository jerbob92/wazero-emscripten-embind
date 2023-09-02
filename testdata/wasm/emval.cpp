#include <emscripten/bind.h>
#include <emscripten/val.h>
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

EMSCRIPTEN_BINDINGS(emval) {
    function("doEmval", &doEmval);
}