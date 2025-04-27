package main // TODO: Implement the security measures and connect them with `secure.go`


import (
	"context"
	// "crypto/sha256"
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
	// timestamp := time.Now().Unix() // Current UNIX time, as a monotonically-increasing UUID
	var wg sync.WaitGroup // Using Go's excellent concurrency support with WaitGroup
	var mu sync.Mutex // Using Go's excellent concurrency support with Mutex
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second) // Setting the timeout limit
	defer cancel() // Setting the time limit for each process

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
					
					res, err := etcdClient.Get(ctx, r.Key) // Getting the data from the etcd server
					if err != nil { // Handling the first error case
						log.Printf("ERROR %v: Failed to read key %s", err, r.Key)
						return
					}
					
					// Putting the data inside the client's interface
					var val interface{}
					if len(res.Kvs) > 0 { // Marshalling the data into the JSON
						log.Printf("Reading the raw: (key = %s, value = %s)", r.Key, string(res.Kvs[0].Value))
						var decoded interface{}
						if err := json.Unmarshal(res.Kvs[0].Value, &decoded); err != nil { // Handling the second error case
							log.Printf("ERROR %v: Failed to unmarshal value for key %s", err, r.Key)
						} else { // Decoded the data into the JSON
							val = decoded
						}
					}

					// Concurrency logic
					mu.Lock()
					cliResp[r.RID] = append(cliResp[r.RID], val)
					mu.Unlock()
				}(req)

			case "write": // Handling a write request
				wg.Add(1) // Adding this request to the Wait Group
				go func(r Request) { // Using an internal helper function inside the read request's given wait group
					defer wg.Done() // Making others wait until we are done

					_, err := etcdClient.Put(ctx, r.Key, string(r.Val)) // Putting the data inside the server's etcd instance
					if err != nil { // Handling the second error case
						log.Printf("Error %v: Failed to write key %s", err, r.Key)
						return
					}
					
					// Concurrency logic
					mu.Lock()
					cliResp[r.RID] = append(cliResp[r.RID], "correct_write")
					mu.Unlock()
				}(req)
			
			default: // Handling all other unsupported edge cases
				log.Printf("Unsupported operation: %s", req.Op)
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

func main() {
	// TODO: PathORAM data structure tests
	// ORAM constructor and destruction
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
	
	// Command-line flags, with default values and changeable by the client
	apiPort := flag.Int("api-port", 5000, "Listening port for the client's Locker API server (By default, 5000)")
	etcdPort := flag.String("etcd-endpoint", "http://localhost:2379", "Listening port for the server's etcd API server (By default, http://localhost:2379)")
	flag.Parse()
	apiAddr := fmt.Sprintf(":%d", *apiPort) // Reading the API port as an Integer and converting it into a String
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { // Set up an HTTP server
		var requests []Request
		if err := json.NewDecoder(r.Body).Decode(&requests); err != nil { // Handling HTTP errors
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		// Setup an etcd client
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

	// Logging information
	log.Println("Locker 2.0 is running on port", *apiPort)
	log.Fatal(http.ListenAndServe(apiAddr, nil)) // The API server is listening on port apiPort
}
