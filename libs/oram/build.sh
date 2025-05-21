#!/bin/bash
# Ismail Ahmed: Builds the dependencies of PathORAM

# Setting the needed properties
set -e
cd "$(dirname "$0")"

# Define=ing the relative paths
PATHORAM_INCLUDE="../../deps/PathORAM/include"
PATHORAM_OBJ="../../deps/PathORAM/obj"

# Compiling the PathORAM implementation
g++ -std=c++11 -fPIC -c -I$PATHORAM_INCLUDE PathORAM.cpp -o PathORAM.o
g++ -std=c++11 -fPIC -c dummy.cpp -o dummy.o

# Linking everything together
ar rcs libpathoram.a PathORAM.o dummy.o $PATHORAM_OBJ/*.o

# Letting the user know that we have built the library
echo "Library built successfully"
