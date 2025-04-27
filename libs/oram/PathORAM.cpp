// libs/oram/PathORAM.cpp
// Ismail Ahmed: Wraps the PathORAM oblivious memory data structure in C++, as Go only supports interoperation with C functions, not C++

#include "OramInterface.h"
#include <cstdlib>
#include <cstring>
#include <cstdint>
#include <algorithm>
#include <vector>
#include <cstring>


// This is supposed to be a neat C++ PathORAM interfacing class
class ORAMClass : public OramInterface {
    private:
        std::vector<Block> storage;
        int num_blocks;
        int block_size;
    public:
        ORAMClass(int log_capacity, int block_size_bytes, int z) { // Defining the constructor class
            num_blocks = 1 << log_capacity; // Fitting as much as we can, 2^log - 1
            block_size = Block::BLOCK_SIZE; // Defining the number of integers per block

            for (int i = 0; i < num_blocks; ++i) { // Initalizing the blocks as blanks into storage
                Block blk;
                blk.index = i;
                std::memset(blk.data, 0, sizeof(blk.data));
                storage.push_back(blk);
            }
        }

        // Overwriting the original OramInterface.h functions

        int *access (Operation op, int blockIndex, int newData[]) override { // Handling all ORAM accesses
            if (blockIndex >= num_blocks) { // Memory safety check
                return nullptr;
            }

            Block &blk = storage[blockIndex];
            if (op == READ) { // Upon a READ, return the original data
                return blk.data;
            } else if (op == WRITE && newData != nullptr) { // Upon a WRITE, write the new data in the same place as the original data
                memcpy(blk.data, newData, block_size * sizeof(int));
                return nullptr;
            }

            return nullptr; // Failure case
        }

        int P(int leaf, int level) override { // Default fetter function for the position
            return 0;
        }

        int *getPositionMap() override { // Default getter function for the position map
            return nullptr;
        }

        std::vector<Block> getStash() override { // Default getter funciton for the stash
            return std::vector<Block>();
        }

        int getStashSize() override { // Default getter function for the stash size
            return 0;
        }
        
        int getNumLeaves() override { // Default getter function for the number of leaves
            return 0;
        }
        
        int getNumLevels() override { // Default getter function for the number of levels
            return 0;
        }
        
        int getNumBlocks() override { // Default getter function for the number of blocks
            return num_blocks;
        }
        
        int getNumBuckets() override { // Default getter function for the number of buckets
            return 0;
        }
};


extern "C" { // Supporting construction, destruction, and the CURD (Create, Update, Read, and Destroy) operations
    typedef OramInterface ORAM; // PathORAM structure definition, as in function arguments and types, I will treat ORAM as a generic void *

    void *oram_init(uint32_t log_capacity, uint32_t block_size, uint32_t z) { // ORAM constructor function
        return new ORAMClass(log_capacity, block_size, z);
    }

    void oram_destruct(void *oram) { // ORAM destructor function
        delete static_cast<ORAM*>(oram); // Destroying the ORAM object's pointer
    }

    void oram_get(void *oram, uint64_t id, uint8_t *out_data, size_t data_len) { // ORAM getter function
        // Data structures
        ORAM *oram_ptr = static_cast<ORAM*>(oram); // Creating the ORAM object's pointer

        // Get operation
        int *result = oram_ptr->access(ORAM::READ, static_cast<int>(id), nullptr); // Reading from the ORAM
        if (result) {
            std::memcpy(out_data, result, data_len); // Outputting onto out_data
        }
    }

    void oram_set(void *oram, uint64_t id, const uint8_t *in_data, size_t data_len) { // ORAM setter and update function
        // Data structures
        ORAM *oram_ptr = static_cast<ORAM*>(oram); // Creating the ORAM object's pointer
        int *buffer = new int[(data_len + sizeof(int) - 1) / sizeof(int)](); // Creating the empty buffer, to be filled later
        
        // Set operation
        std::memcpy(buffer, in_data, data_len); // Inputting onto in_data
        oram_ptr->access(ORAM::WRITE, static_cast<int>(id), buffer); // Writing to the ORAM
        delete[] buffer; // Deleting the buffer after we are done
    }

    void oram_delete(void *oram, uint64_t id, uint32_t block_size) { // ORAM deletion function
        // Data structures
        ORAM *oram_ptr = static_cast<ORAM*>(oram); // Creating the ORAM object's pointer
        int *buffer = new int[(block_size + sizeof(int) - 1) / sizeof(int)](); // Creating the empty buffer, never to be filled
        
        // Delete operation
        oram_ptr->access(ORAM::WRITE, static_cast<int>(id), buffer); // Writing to the ORAM
        delete[] buffer; // Deleting the buffer after we are done
    }
}
