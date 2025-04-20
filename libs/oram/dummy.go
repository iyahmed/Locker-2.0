package oram
/*
#cgo CXXFLAGS: -std=c++17 -fPIC -DSHARED -DNO_STORAGE_ADAPTERS -I${SRCDIR}/../../deps/path-oram/path-oram/include
#cgo LDFLAGS: -L${SRCDIR} -lpathoram -lstdc++
extern int trigger_cgo();
*/
import "C"