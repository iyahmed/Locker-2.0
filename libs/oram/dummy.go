package oram
/*
#cgo CXXFLAGS: -std=c++11 -fPIC -I${SRCDIR}/../../deps/PathORAM/include
#cgo LDFLAGS: -L${SRCDIR} -l:libpathoram.a -L${SRCDIR}/../../deps/PathORAM/obj -lstdc++
extern int trigger_cgo();
*/
import "C"