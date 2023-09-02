#include <emscripten/bind.h>

using namespace emscripten;

const bool SOME_CONSTANT_1 = false;
const float SOME_CONSTANT_2 = 2;
const double SOME_CONSTANT_3 = 3;
const int SOME_CONSTANT_4 = 4;
const std::string SOME_CONSTANT_5 = "TestString";
const char SOME_CONSTANT_6 = 'C';
const long long SOME_CONSTANT_7 = 7;
const unsigned short SOME_CONSTANT_8 = 8;
const unsigned int SOME_CONSTANT_9 = 9;
const unsigned short SOME_CONSTANT_10 = 10;
const unsigned char SOME_CONSTANT_11 = 11;
const unsigned long SOME_CONSTANT_12 = 12;
const std::wstring SOME_CONSTANT_13 = L"TestWideString";
const bool SOME_CONSTANT_14 = true;
const unsigned long long SOME_CONSTANT_15 = 15;

EMSCRIPTEN_BINDINGS(constants) {
    constant("SOME_CONSTANT_1", SOME_CONSTANT_1);
    constant("SOME_CONSTANT_2", SOME_CONSTANT_2);
    constant("SOME_CONSTANT_3", SOME_CONSTANT_3);
    constant("SOME_CONSTANT_4", SOME_CONSTANT_4);
    constant("SOME_CONSTANT_5", SOME_CONSTANT_5);
    constant("SOME_CONSTANT_6", SOME_CONSTANT_6);
    constant("SOME_CONSTANT_7", SOME_CONSTANT_7);
    constant("SOME_CONSTANT_8", SOME_CONSTANT_8);
    constant("SOME_CONSTANT_9", SOME_CONSTANT_9);
    constant("SOME_CONSTANT_10", SOME_CONSTANT_10);
    constant("SOME_CONSTANT_11", SOME_CONSTANT_11);
    constant("SOME_CONSTANT_12", SOME_CONSTANT_12);
    constant("SOME_CONSTANT_13", SOME_CONSTANT_13);
    constant("SOME_CONSTANT_14", SOME_CONSTANT_14);
    constant("SOME_CONSTANT_15", SOME_CONSTANT_15);
}