// libs/oram/PathORAM.go
// Ismail Ahmed: Implements the PathORAM oblivious memory data structure

package oram
/*
#include <stdint.h>
#include <stdlib.h>
void* oram_init(uint32_t log_capacity, uint32_t block_size, uint32_t z);
void  oram_destruct(void* oram);
void  oram_get(void* oram, uint64_t id, uint8_t *out_data, size_t data_len);
void  oram_set(void* oram, uint64_t id, const uint8_t *in_data, size_t data_len);
void  oram_delete(void* oram, uint64_t id, uint32_t block_size);
*/
import "C"
import "unsafe"
// #include "OramInterface.h"
type ORAM struct { // Using the C++ wrapper's ORAM struct for an unsafe C pointer
	ptr unsafe.Pointer
}

func ORAM_Init(logCapacity, blockSize, z uint32) *ORAM { // ORAM constructor function
	ptr := C.oram_init(C.uint32_t(logCapacity), C.uint32_t(blockSize), C.uint32_t(z)); // Constructing the new ORAM object
	
	return &ORAM{ptr: ptr} // Returning the new ORAM object
}

func (o *ORAM) ORAM_Destruct() { // ORAM destructor function
	if o.ptr != nil { // If we can, we will destruct the ORAM object
		C.oram_destruct(o.ptr) // Destructing the ORAM object
		o.ptr = nil // Setting the wild pointer to NIL
	}
}

func (o *ORAM) ORAM_Get(id uint64, blockSize int) []byte { // ORAM getter function
	out := make([]byte, blockSize) // Creating an output object
	C.oram_get(o.ptr, C.uint64_t(id), (*C.uint8_t)(unsafe.Pointer(&out[0])), C.size_t(blockSize)) // Getting the requested value from the ORAM
	
	return out // Returning the output object
}

func (o *ORAM) ORAM_Set(id uint64, data []byte) { // ORAM setter function
	if len(data) == 0 { // If we cannot, we will not even try
		return
	}

	C.oram_set(o.ptr, C.uint64_t(id), (*C.uint8_t)(unsafe.Pointer(&data[0])), C.size_t(len(data))) // Setting the request value to the ORAM
}

func (o *ORAM) ORAM_Delete(id uint64, blockSize uint32) { // ORAM deletion function
	C.oram_delete(o.ptr, C.uint64_t(id), C.uint32_t(blockSize)) // Writing an empty value to the requested entry from the ORAM
}
