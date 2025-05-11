// libs/emm/emm_server.go
// Ismail Ahmed: Implements the server-side of the encrypted mini-map (EMM) oblivious memory data structure

package emm

// Needed imports
import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"locker/libs/oram"
	"locker/libs/hirb"
)

// Structs for the Response and Request JSON objects
type Response struct {
	Result interface{} `json:"result"`
}
type RequestServerSide struct {
	Op  string      `json:"op"`
	Key string      `json:"key"`
	Val interface{} `json:"val,omitempty"`
}

// The default AES-GCM symmetric key is stored server-side
const keyFile = "emm_key.bin"

// Global variables' default states
var (
	oramStore   *oram.ORAM = nil
	idLock      sync.RWMutex
	symKey      []byte
	blockSize   uint32     = 256
	logCapacity uint32     = 6
	z           uint32     = 4
	idCounter   uint64     = 1
)

// Starting the EMM server
func InitServer() {
	symKey = loadOrGenerateKey()
	oramStore = oram.ORAM_Init(logCapacity, blockSize, z)
}

// Stopping the EMM server
func ShutdownServer() {
	if oramStore != nil {
		oramStore.ORAM_Destruct()
	}
}

// Loading or generating the standard AES-GCM symmetric-key encryption used elsewhere
func loadOrGenerateKey() []byte {
	// Attempting to load the key from the default "emm_key.bin" file
	if data, err := os.ReadFile(keyFile); err == nil && len(data) == 32 {
		fmt.Println("Loaded the default AES symmetric key from disk.")
		return data
	} 

	// If we cannot load the key, we will generate a new one and store it inside "emm_key.bin"
	newKey := make([]byte, 32) // Using AES-256 for additional security
	if _, err := rand.Read(newKey); err != nil { // If we could not generate the symmetric key, we must error out
		log.Fatalf("ERROR 0: Could not generate the AES symmetric key: %v", err)
	}
	if err := os.WriteFile(keyFile, newKey, 0600); err != nil { // If we could not generate the symmetric key, we must error out
		log.Fatalf("ERROR 1: Could not write the AES symmetric key into disk: %v", err)
	}

	// Alerting the user that we had to successfuly make a new AES symmetric key
	fmt.Println("Successfully generated and saved a new AES symmetric key into disk")

	return newKey
}

// Encrypting a block of data using ACM-GCM
func encrypt(val []byte) ([]byte, error) {
	block, err := aes.NewCipher(symKey)
	if err != nil { // If at any point the encryption fails, we must error out
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, val, nil), nil
}

// Decrypting a block of data using ACM-GCM
func decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(symKey)
	if err != nil { // If at any point the encryption fails, we must error out
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ERROR 2: Could not decrypt due to too-short ciphertext")
	}
	nonce := ciphertext[:gcm.NonceSize()]
	enc := ciphertext[gcm.NonceSize():]

	return gcm.Open(nil, nonce, enc, nil)
}

// Getting or assigning an oblivious index for logical keys
func getOrCreateIndex(key string) (uint64, error) {
	// Attempting to get the given key's index inside the HIRB
	val, err := hirb.Get(key)
	if err == nil && val != nil {
		if idFloat, okay := val.(float64); okay {
			return uint64(idFloat), nil
		}
	}

	// Standard mutex logic
	idLock.Lock()
	newID := idCounter
	idCounter++
	idLock.Unlock()

	// Attempting to set the given key's index inside the HIRB
	if err := hirb.Set(key, newID); err != nil {
		return 0, fmt.Errorf("ERROR 3: The HIRB set operation failed: %w", err)
	}


	return newID, nil
}

// Handling all HTTP requests in JSON format
func HandleEMMRequest(w http.ResponseWriter, r *http.Request) {
	// Receiving JSON requests
	var req RequestServerSide
	body, err := io.ReadAll(r.Body)
	if err != nil || json.Unmarshal(body, &req) != nil {
		http.Error(w, "HTTP ERROR 500: Invalid request: ", http.StatusBadRequest)
		return
	}

	// Creating the relevant key's HIRB index
	key := req.Key
	idx, err := getOrCreateIndex(key)
	if err != nil {
		http.Error(w, "HTTP ERROR 500: The HIRB has failed: " + err.Error(), http.StatusInternalServerError)
		return
	}

	// Switch-case statement to handle all legal operations (GET/Read, SET/Write, DEL/Remove)
	// NOTE: The PUT/Update operation is unneeded because a working SET should always find the given index that it needs to PUT/Update
	switch req.Op {
		case "get":
			// Handling the GET case, where we must obliviously fetch a value from the ORAM
			ciphertext := oramStore.ORAM_Get(idx, int(blockSize))
			end := bytes.IndexByte(ciphertext, 0)
			if end == -1 { // Handling the corner case
				end = len(ciphertext)
			}
			plaintext, err := decrypt(ciphertext[:end])
			if err != nil { // If we cannot decrypt, we must return NIL to the client
				json.NewEncoder(w).Encode(Response{Result: nil})
				return
			}
			var decoded interface{}
			if err := json.Unmarshal(plaintext, &decoded); err != nil { // If we cannot decrypt, we must return NIL to the client
				json.NewEncoder(w).Encode(Response{Result: nil})
				return
			}
			json.NewEncoder(w).Encode(Response{Result: decoded})
		case "set":
			// Handling the SET case, where we must obliviously write a value to the ORAM
			raw, err := json.Marshal(req.Val)
			if err != nil {
				http.Error(w, "HTTP ERROR 500: Marshal error: ", http.StatusInternalServerError)
				return
			}
			ciphertext, err := encrypt(raw)
			if err != nil {
				http.Error(w, "HTTP ERROR 500: Encryption error: ", http.StatusInternalServerError)
				return
			}
			oramStore.ORAM_Set(idx, ciphertext)
			json.NewEncoder(w).Encode(Response{Result: "OK"})
		case "del":
			// Handling the DEL case, where we must obliviously delete a value from the ORAM
			oramStore.ORAM_Delete(idx, blockSize)
			json.NewEncoder(w).Encode(Response{Result: "Deleted the chosen data as requested"})
		default:
			// Handling all unsupported operations
			http.Error(w, "HTTP 404: REQUESTED OPERATION NOT FOUND (VALID ONES ARE \"SET\", \"GET\", AND \"DEL\")", http.StatusBadRequest)
	}
}

// func main() {
// 	// Performing all non-HTTP operations in order
// 	symKey = loadOrGenerateKey()
// 	oramStore = oram.ORAM_Init(logCapacity, blockSize, z)
// 	defer oramStore.ORAM_Destruct()

// 	// Performing all HTTP operations in order
// 	http.HandleFunc("/", jsonHandler)
// 	fmt.Println("EMM server running on http://localhost:8245")
// 	log.Fatal(http.ListenAndServe(":8245", nil))
// }
