emcc -sERROR_ON_UNDEFINED_SYMBOLS=0 -sEXPORTED_FUNCTIONS="_free,_malloc" -g classes.cpp functions.cpp constants.cpp enums.cpp structs.cpp -o tests.wasm -lembind --no-entry
