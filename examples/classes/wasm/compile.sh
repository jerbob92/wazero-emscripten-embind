emcc -sERROR_ON_UNDEFINED_SYMBOLS=0 -sEXPORTED_FUNCTIONS="_free,_malloc" -g classes.cpp -o classes.wasm -lembind --no-entry
