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
	// Memory safety checks
	if o.ptr == nil || blockSize < 0 {
		return nil
	}
	
	buff := make([]byte, blockSize) // Creating an output buffer
	C.oram_get(o.ptr, C.uint64_t(id), (*C.uint8_t)(unsafe.Pointer(&buff[0])), C.size_t(blockSize)) // Getting the requested value from the ORAM
	
	return buff // Returning the output object
}

func (o *ORAM) ORAM_Set(id uint64, data []byte) { // ORAM setter function
	// Memory safety checks
	if o.ptr == nil || len(data) == 0 { // If we cannot, we will not even try
		return
	}
	
	// Trivial padding to the next power of four for memory safety reasons
	padding := (4 - (len(data) % 4)) % 4 // Using modolus for trivial padding math
	padded := append(data, make([]byte, padding)...) // Padding our the data

	C.oram_set(o.ptr, C.uint64_t(id), (*C.uint8_t)(unsafe.Pointer(&padded[0])), C.size_t(len(padded))) // Setting the request value to the ORAM
}

func (o *ORAM) ORAM_Delete(id uint64, blockSize uint32) { // ORAM deletion function
	if o.ptr == nil { // Memory safety check
		return
	}

	C.oram_delete(o.ptr, C.uint64_t(id), C.uint32_t(blockSize)) // Writing an empty value to the requested entry from the ORAM
}
