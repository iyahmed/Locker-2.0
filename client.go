package main


import (
	"context"
	// "crypto/sha256"
	"encoding/json"
	// "fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	clientv3 "go.etcd.io/etcd/client/v3"
	// "locker/libs"
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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { // Set up an HTTP server
		var requests []Request
		if err := json.NewDecoder(r.Body).Decode(&requests); err != nil { // Handling HTTP errors
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Setup an etcd client
		etcdClient, err := clientv3.New(clientv3.Config{
			Endpoints:   []string{"http://localhost:2379"},
			DialTimeout: 5 * time.Second,
		})
		if err != nil { // Handling etcd timeouts
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer etcdClient.Close()

		// Processing etcd responses to the API
		responses := handleRequests(requests, etcdClient)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(responses); err != nil { // Handling etcd errors
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Logging information
	log.Println("Proxy running on port 5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
