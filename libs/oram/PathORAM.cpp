// libs/oram/PathORAM.cpp
// Ismail Ahmed: Wraps the PathORAM oblivious memory data structure in C++, as Go only supports interoperation with C functions, not C++

#include "oram.hpp"
#include <cstdlib>
#include <cstring>
#include <cstdint>
#include <algorithm>

extern "C" { // Supporting construction, destruction, and the CURD (Create, Update, Read, and Destroy) operations
    typedef PathORAM::ORAM ORAM; // PathORAM structure definition, as in function arguments and types, I will treat ORAM as a generic void *

    void *oram_init(uint32_t log_capacity, uint32_t block_size, uint32_t z) { // ORAM constructor function
        return new ORAM(log_capacity, block_size, z);
    }

    void oram_destruct(void *oram) { // ORAM destructor function
        // Data structures
        ORAM *oram_ptr = (ORAM *) oram; // Creating the ORAM object's pointer

        // Destruction operation
        delete oram_ptr; // Destroying the ORAM object's pointer
    }

    void oram_get(void *oram, uint64_t id, uint8_t *out_data, size_t data_len) { // ORAM getter function
        // Data structures
        ORAM *oram_ptr = (ORAM *) oram; // Creating the ORAM object's pointer
        PathORAM::bytes output; // Creating the returned output

        // Get operation
        oram_ptr->get(id, output); // Using the original function
        size_t copy_len = std::min(data_len, output.size()); // Minimizing our data usage
        std::memcpy(out_data, output.data(), copy_len); // Outputting onto out_data
    }

    void oram_set(void *oram, uint64_t id, const uint8_t *data, size_t data_len) { // ORAM setter and update function
        // Data structures
        ORAM *oram_ptr = (ORAM *) oram; // Creating the ORAM object's pointer
        PathORAM::bytes input(data, data + data_len); // Acommodating our new value size
        
        // Set operation
        oram_ptr->put(id, input); // Using the original function
    }

    void oram_delete(void *oram, uint64_t id, uint32_t block_size) { // ORAM deletion function
        // Data structures
        ORAM *oram_ptr = (ORAM *) oram; // Creating the ORAM object's pointer
        PathORAM::bytes dummy(block_size, 0); // Creating a dummy value that has nothing, as a empty/nil value
        
        // Delete operation
        oram_ptr->put(id, dummy); // Using the original function
    }
}
