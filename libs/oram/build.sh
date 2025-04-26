#!/bin/bash
# Ismail Ahmed: Builds the dependencies of PathORAM

# Set properties
set -e
cd "$(dirname "$0")"

# Define paths
PATHORAM_INCLUDE="../../deps/PathORAM/include"
PATHORAM_OBJ="../../deps/PathORAM/obj"

# Compile the  implementation
g++ -std=c++11 -fPIC -c -I$PATHORAM_INCLUDE PathORAM.cpp -o PathORAM.o
g++ -std=c++11 -fPIC -c dummy.cpp -o dummy.o

# Link everything together
ar rcs libpathoram.a PathORAM.o dummy.o $PATHORAM_OBJ/*.o

# Clean up
echo "Library built successfully"
