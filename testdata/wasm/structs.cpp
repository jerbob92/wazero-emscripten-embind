#include <emscripten/bind.h>

using namespace emscripten;

struct Point2f {
    float x;
    float y;
};

// Array fields are treated as if they were std::array<type,size>
struct ArrayInStructStruct {
    int field[2];
};

struct PersonRecord {
    std::string name;
    int age;
    ArrayInStructStruct structArray;
};

PersonRecord findPersonAtLocation(Point2f)  {
    return PersonRecord{
      .name="123",
      .age=12,
      .structArray=ArrayInStructStruct{
        .field={1,2}
      },
    };
}

void setPersonAtLocation(Point2f, PersonRecord)  {
}

EMSCRIPTEN_BINDINGS(structs) {
   value_array<Point2f>("Point2f")
       .element(&Point2f::x)
       .element(&Point2f::y)
       ;

  value_object<PersonRecord>("PersonRecord")
      .field("name", &PersonRecord::name)
      .field("age", &PersonRecord::age)
      .field("structArray", &PersonRecord::structArray)
      ;

  value_object<ArrayInStructStruct>("ArrayInStructStruct")
      .field("field", &ArrayInStructStruct::field) // Need to register the array type
      ;

  // Register std::array<int, 2> because ArrayInStructStruct::field is interpreted as such
  // Already registered in embind_test.cpp
  //value_array<std::array<int, 2>>("array_int_2")
  //    .element(emscripten::index<0>())
  //    .element(emscripten::index<1>())
  //    ;

  function("findPersonAtLocation", &findPersonAtLocation);

  function("setPersonAtLocation", &setPersonAtLocation);
}