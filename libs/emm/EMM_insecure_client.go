// libs/emm/emm_insecure_client.go
// Ismail Ahmed: Implements the client-side of the encrypted mini-map (EMM) oblivious memory data structure as an internal library
// NOTE: THIS IS NOT SECURE BECAUSE THE CLIENT AND THE SERVER HAVE ACCESS TO THE SAME INFORMATION, IT SHOULD ONLY USED FOR EMM TESTING AND BENCHMARKING

package emm

// Needed imports
import (
	"fmt"
	"io"
	"os"
	"bytes"
	"log"
	"sync"
	"syscall"
	"time"
	"path/filepath"
	"encoding/json"
	"crypto/cipher"
	"crypto/aes"
	"crypto/rand"
	"locker/libs/oram"
	"locker/libs/hirb"
)

// Helper struct
type Client struct {
	oram      *oram.ORAM
	symKey    []byte
	idCounter uint64
	idLock    sync.RWMutex
	tempDir   string
}

// Global variables' default states
const (
	logCap    = uint32(6)
)

// New EMM client interface function

// Constructor for the internal client constructor
func NewClient() *Client {
	key := newLoadOrGenerateKey()
	// Creating a temporary file for the Baseline EMM's disk read/writes
	tempDir := "/mnt/data"
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		tempDir = os.TempDir()
	}
	if err := os.MkdirAll(tempDir, 0700); err != nil {
		log.Fatalf("Could not create temp directory: %v", err)
	}
	return &Client{
		oram:      oram.ORAM_Init(logCap, blockSize, z),
		symKey:    key,
		idCounter: 1,
		tempDir:   tempDir,
	}
}

// EMM server implementation functions

// Loading or generating the standard AES-GCM symmetric-key encryption used elsewhere
func newLoadOrGenerateKey() []byte {
	// Attempting to load the key from the default "emm_key.bin" file
	if data, err := os.ReadFile(keyFile); err == nil && len(data) == 32 {
		fmt.Println("Loaded the default AES symmetric key from disk.\n")
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
func (c *Client) newEncrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.symKey)
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

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// Decrypt data using AES-GCM
func (c *Client) newDecrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.symKey)
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
func (c *Client) newGetOrCreateIndex(key string) (uint64, error) {
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

// Public newClient EMM API functions

// newClient's replacement for the HTTP PUT/Update operation
func (c *Client) Put(key string, val string) error {
	// Obliviously accessing the HIRB to find the correct element, if possible
	idx, err := c.newGetOrCreateIndex(key)
	if err != nil {
		return err
	}
	// Marshalling the JSON data that is being set into the HIRB, if possible
	raw, err := json.Marshal([]string{val})
	if err != nil {
		return err
	}
	// Encrypting the value that is about to be sent into the HIRB, if possible
	ciphertext, err := c.newEncrypt(raw)
	if err != nil {
		return err
	}
	// Sending the new (key, value) pair into the ORAM
	c.oram.ORAM_Set(idx, ciphertext)
	
	// Writing it into a temporary file
	path := filepath.Join(c.tempDir, fmt.Sprintf("put_%d_%s", time.Now().UnixNano(), key))
	f, err := os.Create(path)
	if err == nil {
		pad := make([]byte, 4096)
		copy(pad, ciphertext)
		_, _ = f.Write(pad)
		_ = f.Sync()
		_ = f.Close()
		syscall.Sync()
	}

	// Returning NIL upon success
	return nil
}

// newClient's replacement for the HTTP GET/Read operation
func (c *Client) Get(key string) ([]string, error) {
	// Obliviously accessing the HIRB to find the correct element, if possible
	idx, err := c.newGetOrCreateIndex(key)
	if err != nil {
		return nil, err
	}
	// Reading the encrypted ciphertext, if possible
	ciphertext := c.oram.ORAM_Get(idx, int(blockSize))
	end := bytes.IndexByte(ciphertext, 0)
	if end == -1 {
		end = len(ciphertext)
	}
	// Decrypting the value that is found from the HIRB, if possible
	plaintext, err := c.newDecrypt(ciphertext[:end])
	if err != nil {
		return nil, nil // This is to be consistent with the corresponding secure EMM implementation
	}
	// Unmarshalling the JSON data that is being read from the HIRB, if possible
	var result []string
	if err := json.Unmarshal(plaintext, &result); err != nil && err != io.EOF {
		return nil, nil // This is to be consistent with the corresponding secure EMM implementation
	}
	
	// Writing it into a temporary file
	path := filepath.Join(c.tempDir, fmt.Sprintf("get_%d_%s", time.Now().UnixNano(), key))
	f, err := os.Create(path)
	if err == nil {
		pad := make([]byte, 4096)
		copy(pad, plaintext)
		_, _ = f.Write(pad)
		_ = f.Sync()
		_ = f.Close()
		syscall.Sync()
	}

	// Returning either the found value upon success or NIL upon failure
	return result, nil
}

// newClient's replacement for the HTTP DEL/Remove operation
func (c *Client) Delete(key string) error {
	// Obliviously accessing the HIRB to find the correct element, if possible
	idx, err := c.newGetOrCreateIndex(key)
	if err != nil {
		return err
	}

	// Deleting the value that is found in the ORAM
	c.oram.ORAM_Delete(idx, blockSize)
	
	// Writing it into a temporary file
	path := filepath.Join(c.tempDir, fmt.Sprintf("delete_%d_%s", time.Now().UnixNano(), key))
	f, err := os.Create(path)
	if err == nil {
		pad := make([]byte, 4096)
		copy(pad, []byte(key))
		_, _ = f.Write(pad)
		_ = f.Sync()
		_ = f.Close()
		syscall.Sync()
	}

	// Returning NIL upon success
	return nil
}
