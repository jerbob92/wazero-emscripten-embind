#include <emscripten/bind.h>
using namespace emscripten;

bool bool_return_true() {
    return true;
}

bool bool_return_false() {
    return false;
}

void float_return_void(float a) {}

float float_return_float(float a) {
    return a*a;
}

double double_return_double(double a) {
    return a*a;
}

int int_return_int(int a) {
    return a * a;
}

char char_return_char(char a) {
    return a;
}

long long_return_long(long a) {
    return a+a;
}

short short_return_short(short a) {
    return a+a;
}

unsigned char uchar_return_uchar(unsigned char a) {
    return a;
}

unsigned int uint_return_uint(unsigned int a) {
    return a + a;
}

unsigned long ulong_return_ulong(unsigned long a) {
    return a + a;
}

unsigned short ushort_return_ushort(unsigned short a) {
    return a + a;
}

long long longlong_return_longlong(long long a) {
    return a + a;
}

std::string std_string_return_std_string(std::string in) {
    return "Hello there " + in;
}

std::wstring std_wstring_return_std_wstring(std::wstring in) {
    return L"Hello there " + in;
}

std::vector<int> return_vector () {
  std::vector<int> v(10, 1);
  return v;
}

std::map<int, std::string> return_map () {
  std::map<int, std::string> m;
  m.insert(std::pair<int, std::string>(10, "This is a string."));
  return m;
}

char data_char[] = {0, 1, 2, 3, 4, 5};
unsigned char data_unsigned_char[] = {0, 1, 2, 3, 4, 5};
int data_int[] = {0, 1, 2, 3, 4, 5};
unsigned int data_unsigned_int[] = {0, 1, 2, 3, 4, 5};
long data_long[] = {0, 1, 2, 3, 4, 5};
unsigned long data_unsigned_long[] = {0, 1, 2, 3, 4, 5};
short data_short[] = {0, 1, 2, 3, 4, 5};
unsigned short data_unsigned_short[] = {0, 1, 2, 3, 4, 5};
long long data_longlong[] = {0, 1, 2, 3, 4, 5};
unsigned long long data_unsigned_longlong[] = {0, 1, 2, 3, 4, 5};
double data_double[] = {0, 1, 2, 3, 4, 5};
float data_float[] = {0, 1, 2, 3, 4, 5};

val get_memory_view_char() {
    return val(typed_memory_view(sizeof data_char, data_char));
}

val get_memory_view_unsigned_char() {
    return val(typed_memory_view(sizeof data_unsigned_char, data_unsigned_char));
}

val get_memory_view_int() {
    return val(typed_memory_view(sizeof data_int, data_int));
}

val get_memory_view_unsigned_int() {
    return val(typed_memory_view(sizeof data_unsigned_int, data_unsigned_int));
}

val get_memory_view_long() {
    return val(typed_memory_view(sizeof data_long, data_long));
}

val get_memory_view_unsigned_long() {
    return val(typed_memory_view(sizeof data_unsigned_long, data_unsigned_long));
}

val get_memory_view_short() {
    return val(typed_memory_view(sizeof data_short, data_short));
}

val get_memory_view_unsigned_short() {
    return val(typed_memory_view(sizeof data_unsigned_short, data_unsigned_short));
}

val get_memory_view_longlong() {
    return val(typed_memory_view(sizeof data_longlong, data_longlong));
}

val get_memory_view_unsigned_longlong() {
    return val(typed_memory_view(sizeof data_unsigned_longlong, data_unsigned_longlong));
}

val get_memory_view_double() {
    return val(typed_memory_view(sizeof data_double, data_double));
}

val get_memory_view_float() {
    return val(typed_memory_view(sizeof data_float, data_float));
}

EMSCRIPTEN_BINDINGS(functions) {
    function("bool_return_true", &bool_return_true);
    function("bool_return_false", &bool_return_false);
    function("float_return_void", &float_return_void);
    function("float_return_float", &float_return_float);
    function("double_return_double", &double_return_double);
    function("int_return_int", &int_return_int);
    function("char_return_char", &char_return_char);
    function("long_return_long", &long_return_long);
    function("short_return_short", &short_return_short);
    function("uchar_return_uchar", &uchar_return_uchar);
    function("uint_return_uint", &uint_return_uint);
    function("ulong_return_ulong", &ulong_return_ulong);
    function("ushort_return_ushort", &ushort_return_ushort);
    function("longlong_return_longlong", &longlong_return_longlong);
    function("std_string_return_std_string", &std_string_return_std_string);
    function("std_wstring_return_std_wstring", &std_wstring_return_std_wstring);

    register_vector<int>("vector<int>");
    register_map<int, std::string>("map<int, string>");

    function("return_vector", &return_vector);
    function("return_map", &return_map);

    function("get_memory_view_char", &get_memory_view_char);
    function("get_memory_view_unsigned_char", &get_memory_view_unsigned_char);
    function("get_memory_view_int", &get_memory_view_int);
    function("get_memory_view_unsigned_int", &get_memory_view_unsigned_int);
    function("get_memory_view_long", &get_memory_view_long);
    function("get_memory_view_unsigned_long", &get_memory_view_unsigned_long);
    function("get_memory_view_short", &get_memory_view_short);
    function("get_memory_view_unsigned_short", &get_memory_view_unsigned_short);
    function("get_memory_view_longlong", &get_memory_view_longlong);
    function("get_memory_view_unsigned_longlong", &get_memory_view_unsigned_longlong);
    function("get_memory_view_double", &get_memory_view_double);
    function("get_memory_view_float", &get_memory_view_float);
}