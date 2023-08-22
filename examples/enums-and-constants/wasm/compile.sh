emcc -sERROR_ON_UNDEFINED_SYMBOLS=0 -sEXPORTED_FUNCTIONS="_free,_malloc" -g enums_and_constants.cpp -o enums_and_constants.wasm -lembind --no-entry
