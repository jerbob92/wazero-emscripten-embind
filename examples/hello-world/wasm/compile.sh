emcc -sERROR_ON_UNDEFINED_SYMBOLS=0 -sEXPORTED_FUNCTIONS="_free,_malloc" -g hello_world.cpp -o hello_world.wasm -lembind --no-entry
