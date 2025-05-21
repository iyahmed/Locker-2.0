package main


import (
	"encoding/json"
	"fmt"
	"flag"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	"bytes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"locker/libs/oram"
	"locker/libs/hirb"
	"locker/libs/emm"
)


// Helper structs
type Request struct {
	RID string          `json:"rid"`
	Op  string          `json:"op"`
	Key string          `json:"key"`
	Val json.RawMessage `json:"val"`
}


func handleRequests(requests []Request, etcdClient *clientv3.Client) map[string]interface{} {
	// Function variables
	cliResp := make(map[string][]interface{}) // Map of all client responses
	var wg sync.WaitGroup // Using Go's excellent concurrency support with WaitGroup
	var mu sync.Mutex // Using Go's excellent concurrency support with Mutex

	// Process requests
	for _, req := range requests { // Concurrently processing every individual request
		log.Printf("Processing request: %v", req)

		if req.Key == "" { // Skip past blank requests
			continue
		}

		switch strings.ToLower(req.Op) {
			case "read": // Handling a read request
				wg.Add(1) // Adding this request to the Wait Group
				go func(r Request) { // Using an internal helper function inside the read request's given wait group
					defer wg.Done() // Making others wait until we are done

					// SECURE etcd communication
					var decoded interface{}
					val, err := emm.Get(r.Key) // Getting the data from the etcd server
					if err != nil { // Handling the first error case
						log.Printf("ERROR %v: Failed to read key %s", err, r.Key)
						return
					}
					
					// Putting the data inside the client's interface
					switch v := val.(type) { // Using a switch-case to handle all possible data retrevial cases
						case string: // Handling a string return from the etcd server
							if err := json.Unmarshal([]byte(v), &decoded); err != nil {
								log.Printf("ERROR %v: Failed to unmarshal value for key %s", err, r.Key)
								decoded = v // If the unmarshaling fails, then we will return the raw string
							}
						case []byte: // Handling a raw byte return from the etcd server
							if err := json.Unmarshal(v, &decoded); err != nil {
								log.Printf("ERROR %v: Failed to unmarshal value for key %s", err, r.Key)
								decoded = string(v) // If the unmarshaling fails, then we will return the raw string
							}
						default: // Handling all other unsupported edge cases
							if val == nil {
								log.Printf("INFO: Key %s was deleted or not found.", r.Key)
							} else {
								log.Printf("SPECIAL ERROR: Retrieved non-string data type for key %s", r.Key)
							}
							return
					}

					// Concurrency logic
					mu.Lock()
					cliResp[r.RID] = append(cliResp[r.RID], decoded)
					mu.Unlock()
				}(req)

			case "write": // Handling a write request
				wg.Add(1) // Adding this request to the Wait Group
				go func(r Request) { // Using an internal helper function inside the write request's given wait group
					defer wg.Done() // Making others wait until we are done

					// SECURE etcd communication
					if err := emm.Set(r.Key, string(r.Val)); err != nil {
						log.Printf("ERROR %v: Failed to write (key %s, value %s)", err, r.Key, r.Val)
						return
					}
					
					// Concurrency logic
					mu.Lock()
					cliResp[r.RID] = append(cliResp[r.RID], "correct_write")
					mu.Unlock()
				}(req)
		
			case "delete": // Handling a delete request
				wg.Add(1) // Adding this request to the Wait Group
				go func (r Request) { // Using an internal helper function inside the delete request's given wait group
					defer wg.Done() // Making others wait until we are done

					// SECURE etcd communication
					if err := emm.Delete(r.Key); err != nil {
						log.Printf("ERROR %v: Failed to delete key %s", err, r.Key)
						return
					}

					// Concurrency logic
					mu.Lock()
					cliResp[r.RID] = append(cliResp[r.RID], "correct_delete")
					mu.Unlock()
				}(req)

			default: // Handling all other unsupported edge cases
				log.Printf("Unsupported etcd client operation: %s", req.Op)
		}
	}
	wg.Wait() // We will wait until all client-server communication has finished before finalizing our requests

	// Convert response lists to interface{} for returning
	result := make(map[string]interface{})
	for rid, vals := range cliResp {
		if len(vals) == 1 {
			result[rid] = vals[0]
		} else {
			result[rid] = vals
		}
	}

	return result
}


// The ORAM testing function
func testORAM() {
	// Checking if the ORAM behaves like an oblivious map/dictionary with CRUD (Create, Read, Update, and Destroy) Operations
	blockSize, logCap, z := uint32(32), uint32(5), uint32(3) // By descending order
	o := oram.ORAM_Init(logCap, blockSize, z) // Constructing the PathORAM Go object
	defer o.ORAM_Destruct() // For memory safety, destructing the object after we are done using it

	// ORAM CRUD (Create, Read, Update, and Destroy) Operations
	
	// Test 1: Insert "Hello, World!" at index 1
	dataOne := []byte("Hello, World!") // An example of a message that we want to keep hidden
	o.ORAM_Set(1, dataOne) // Putting the first hidden message in index 1
	
	// Test 2: Read "Hello, World!" from index 1
	resultOne := o.ORAM_Get(1, int(blockSize)) // Retrieving the hidden message in index 1
	resultOneLen := bytes.IndexByte(resultOne, 0) // Cutting off the garbage characters at the end
	if resultOneLen == -1 { // Handling the case where all bytes are used
		resultOneLen = len(resultOne)
	}
	fmt.Println("Retrieved from the ORAM at index 1: ", string(resultOne[:resultOneLen])) // Printing out the message at index 1
	
	// Test 3: Update "Hello, World!" at index 1 to be "Bonjour, Monde!"
	dataTwo := []byte("Bonjour, Monde!") // An example of a message that we want to keep hidden
	o.ORAM_Set(1, dataTwo) // Putting the first hidden message in index 1

	// Test 4: Read "Bonjour, Monde!" from index 1
	resultTwo := o.ORAM_Get(1, int(blockSize)) // Retrieving the hidden message in index 1
	resultTwoLen := bytes.IndexByte(resultTwo, 0) // Cutting off the garbage characters at the end
	if resultTwoLen == -1 { // Handling the case where all bytes are used
		resultTwoLen = len(resultTwo)
	}
	fmt.Println("Retrieved from the ORAM at index 1: ", string(resultTwo[:resultTwoLen])) // Printing out the message at index 1

	// Test 5: Insert "Hello, World!" at index 2
	dataThree := []byte("Hello, World!") // An example of a message that we want to keep hidden
	o.ORAM_Set(2, dataThree) // Putting the first hidden message in index 2
	
	// Test 6: Read "Hello, World!" from index 1 and "Bonjour, Monde!" from index 2
	resultThree := o.ORAM_Get(1, int(blockSize)) // Retrieving the hidden message in index 1
	resultThreeLen := bytes.IndexByte(resultThree, 0) // Cutting off the garbage characters at the end
	if resultThreeLen == -1 { // Handling the case where all bytes are used
		resultThreeLen = len(resultThree)
	}
	fmt.Println("Retrieved from the ORAM at index 1: ", string(resultThree[:resultThreeLen])) // Printing out the message at index 1
	resultFour := o.ORAM_Get(2, int(blockSize)) // Retrieving the hidden message in index 2
	resultFourLen := bytes.IndexByte(resultFour, 0) // Cutting off the garbage characters at the end
	if resultFourLen == -1 { // Handling the case where all bytes are used
		resultFourLen = len(resultFour)
	}
	fmt.Println("Retrieved from the ORAM at index 2: ", string(resultFour[:resultFourLen])) // Printing out the message at index 2

	// Test 7: Destroy "Bonjour, Monde!" from index 2
	o.ORAM_Delete(2, blockSize) // Deleting the hidden message in index 2

	// Test 8: Read "Hello, World!" from index 1 and NIL from index 2
	resultFive := o.ORAM_Get(1, int(blockSize)) // Retrieving the hidden message in index 1
	resultFiveLen := bytes.IndexByte(resultFive, 0) // Cutting off the garbage characters at the end
	if resultFiveLen == -1 { // Handling the case where all bytes are used
		resultFiveLen = len(resultFive)
	}
	fmt.Println("Retrieved from the ORAM at index 1: ", string(resultFive[:resultFiveLen])) // Printing out the message at index 1
	resultSix := o.ORAM_Get(2, int(blockSize)) // Retrieving the hidden message in index 2
	resultSixLen := bytes.IndexByte(resultSix, 0) // Cutting off the garbage characters at the end
	if resultSixLen == -1 { // Handling the case where all bytes are used
		resultSixLen = len(resultOne[:resultSixLen])
	}
	fmt.Println("Retrieved from the ORAM at index 2: ", string(resultSix)) // Printing out the message at index 2
}


// The HIRB testing function
func testHIRB() {
	// Checking if the HIRB tree behaves like an oblivious map/dictionary with CRUD (Create, Read, Update, and Destroy) Operations
	keyOne := "English"
	keyTwo := "French"
	valOne := "Hello, World!"
	valTwo := "Bonjour, Monde!"

	// Test 1: Writing the key-value pair "(English, Hello, World!)"
	fmt.Printf("\nTest 1: Writing the key-value pair (English, Hello, World!)\n")
	if err := hirb.Set(keyOne, valOne); err != nil {
		fmt.Printf("HIRB Set failed: %v\n", err)
	}

	// Test 2: Reading the value "Hello, World!" given the key "English"
	fmt.Printf("\nTest 2: Reading the value 'Hello, World!'' given the key 'English'\n")
	res, err := hirb.Get(keyOne)
	if err != nil {
		fmt.Printf("HIRB Get failed: %v\n", err)
	} else {
		fmt.Printf("HIRB Retrieved value: %v\n", res)
	}

	// Test 3: Deleting the key-value pair "(English, Hello, World!)"
	fmt.Printf("\nTest 3: Deleting the key-value pair '(English, Hello, World!)'\n")
	if err := hirb.Delete(keyOne); err != nil {
		fmt.Printf("HIRB Delete failed: %v\n", err)
	}

	// Test 4: Reading the now nonexistent key-value pair "(English, Hello, World!)"
	fmt.Printf("\nTest 4: Reading the now nonexistent key-value pair '(English, Hello, World!)'\n")
	res, err = hirb.Get(keyOne)
	if err != nil {
		fmt.Printf("HIRB Delete succeeded: %v\n", err)
	} else {
		fmt.Printf("HIRB Delete failed: %v\n", res)
	}

	// Test 5: Writing the value "Bonjour, Monde!" given the key "French"
	fmt.Printf("\nTest 5: Writing the value 'Bonjour, Monde!' given the key 'French'\n")
	if err := hirb.Set(keyTwo, valTwo); err != nil {
		fmt.Printf("HIRB Set failed: %v\n", err)
	}

	// Test 6: Writing the value "Hello, World!" given the key "English"
	fmt.Printf("\nTest 6: Writing the value 'Hello, World!' given the key 'English'\n")
	if err := hirb.Set(keyOne, valOne); err != nil {
		fmt.Printf("HIRB Set failed: %v\n", err)
	}

	// Test 7: Reading the value "Hello, World!" given the key "English"
	fmt.Printf("\nTest 7: Reading the value 'Hello, World!' given the key 'English'\n")
	res, err = hirb.Get(keyOne)
	if err != nil {
		fmt.Printf("HIRB Get failed: %v\n", err)
	} else {
		fmt.Printf("HIRB Retrieved value: %v\n", res)
	}

	// Test 8: Reading the value "Bonjour, Monde!" given the key "French"
	fmt.Printf("\nTest 7: Reading the value 'Bonjour, Monde!' given the key 'French'\n")
	res, err = hirb.Get(keyTwo)
	if err != nil {
		log.Printf("HIRB Get failed: %v\n", err)
	} else {
		log.Printf("HIRB Retrieved value: %v\n", res)
	}
}


// The EMM testing function
func testEMM() {
	// Checking if the HIRB tree behaves like an oblivious map/dictionary with CRUD (Create, Read, Update, and Destroy) Operations
	keyOne   := "username"
	keyTwo   := "password"
	valOne   := "Alice"
	valTwo   := "MyHiddenPassword"
	valThree := "MySecretPassword"

	// EMM CRUD (Create, Read, Update, and Destroy) Operations
	
	// Test 1: Writing the pair "(username, Alice)"
	fmt.Println("\nTest 1: Writing the pair '(username, Alice)'\n")
	if err := emm.Set(keyOne, valOne); err != nil {
		fmt.Printf("EMM Set failed: %v\n", err)
	}

	// Test 2: Writing the pair "(password, MyHiddenPassword)"
	fmt.Println("\nTest 2: Writing the pair '(password, MyHiddenPassword)'\n")
	if err := emm.Set(keyTwo, valTwo); err != nil {
		fmt.Printf("EMM Set failed: %v\n", err)
	}

	// Test 3: Reading the value "Alice" given the key "username"
	fmt.Println("\nTest 3: Reading the value 'Alice' given the key 'username'\n")
	res1, err := emm.Get(keyOne)
	if err != nil {
		fmt.Printf("EMM Get failed: %v\n", err)
	} else {
		fmt.Printf("EMM Retrieved value: %v\n", res1)
	}

	// Test 4: Updating the pair "(password, MySecretPassword)"
	fmt.Println("\nTest 4: Updating the pair '(password, MySecretPassword)'\n")
	if err := emm.Set(keyTwo, valThree); err != nil {
		fmt.Printf("EMM Set failed: %v\n", err)
	}

	// Test 5: Reading the value "MySecretPassword" given the key "password"
	fmt.Println("\nTest 5: Reading the value 'MySecretPassword' given the key 'password'\n")
	res2, err := emm.Get(keyTwo)
	if err != nil {
		fmt.Printf("EMM Get failed: %v\n", err)
	} else {
		fmt.Printf("EMM Retrieved value: %v\n", res2)
	}

	// Test 6: Deleting the pair "(password, MySecretPassword)"
	fmt.Println("\nTest 6: Deleting the pair '(password, MySecretPassword)'\n")
	if err := emm.Delete(keyTwo); err != nil {
		fmt.Printf("EMM Delete failed: %v\n", err)
	}

	// Test 7: Reading the value NIL given the key "password"
	fmt.Println("\nTest 7: Reading the value NIL given the key 'password'\n")
	res3, err := emm.Get(keyTwo)
	if err != nil {
		fmt.Printf("EMM Get failed: %v\n", err)
	} else {
		fmt.Printf("EMM Retrieved value (should be NIL): %v\n", res3)
	}
}


func main() {
	// All test functions are disabled by default, although they can be enabled easily by uncommenting out their lines
	// testORAM() // Testing the PathORAM Go interface
	// testHIRB() // Testing the vORAM+HIRB Go interface
	// 
	emm.InitServer()
	defer emm.ShutdownServer()
	http.HandleFunc("/emm", emm.HandleEMMRequest)
	go func() { // Using Go's goroutines to run the EMM's HTTP server in the background to allow for the tests to run
		log.Println("EMM server is running on http://localhost:8245")
		log.Fatal(http.ListenAndServe(":8245", nil))
	}()
	// testEMM() // Testing the EMM Go implementation
	time.Sleep(1 * time.Second) // 1-second sleep call to finish all async operations

	// Command-line flags, with default values and changeable by the client
	apiPort := flag.Int("api-port", 5000, "Listening port for the client's Locker API server (By default, 5000/etcd)")
	etcdPort := flag.String("etcd-endpoint", "http://localhost:2379", "Listening port for the server's etcd API server (By default, http://localhost:2379)")
	flag.Parse()
	apiAddr := fmt.Sprintf(":%d", *apiPort) // Reading the API port as an Integer and converting it into a String
	
	http.HandleFunc("/etcd", func(w http.ResponseWriter, r *http.Request) { // Setting up an HTTP server for all usage of Locker 2.0 (default port: 5000/etcd)
		var requests []Request
		if err := json.NewDecoder(r.Body).Decode(&requests); err != nil { // Handling HTTP errors
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		// Setting up an etcd client
		etcdClient, err := clientv3.New(clientv3.Config{
			Endpoints:   []string{*etcdPort},
			DialTimeout: 5 * time.Second,
		}) // The etcd server is listening on port etcdPort
		if err != nil { // Handling etcd timeouts
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer etcdClient.Close() // We do not close the etcd server until we have received all the responses back

		// Processing etcd responses to the API
		responses := handleRequests(requests, etcdClient)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(responses); err != nil { // Handling etcd errors
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Logging information out onto console
	log.Println("Locker 2.0 is running on port", *apiPort)
	log.Fatal(http.ListenAndServe(apiAddr, nil)) // The API server is listening on port apiPort
}
