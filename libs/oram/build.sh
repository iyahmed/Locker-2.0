#!/bin/bash

# libs/build.sh
# Ismail Ahmed: Builds all the C++ wrappers of pre-existing oblivious data structure implementation, such as PathORAM

# Setting the environment variables
set -e
# Getting the directory of the libs folder
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT="$DIR/../../deps/path-oram/path-oram"
SRC="$ROOT/src"
INCLUDE="$ROOT/include"

g++ -std=c++17 -fPIC \
    -I "$INCLUDE" \
    -DSHARED -DNO_STORAGE_ADAPTERS \
    -Wall -O3 \
    -shared -o "$DIR/libpathoram.so" \
    "$DIR/PathORAM.cpp" \
    "$SRC/oram.cpp" \
    "$SRC/utility.cpp" \
    # "$SRC/config.cpp" \
    # "$SRC/logger.cpp"
