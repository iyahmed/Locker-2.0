// libs/emm/emm_client.go
// Ismail Ahmed: Implements the client-side of the encrypted mini-map (EMM) oblivious memory data structure

package emm

// Needed imports
import (
	"encoding/json"
	"net/http"
	"bytes"
	"fmt"
	"io"
)

// Gloabl variable for the EMM server's port
const endPoint = "http://localhost:8245/emm"

// Structs for the Response and Request JSON objects
// type Response struct {
// 	Result interface{} `json:"result"`
// }
// type Request struct {
// 	Op  string      `json:"op"`
// 	Key string      `json:"key"`
// 	Val interface{} `json:"val,omitempty"`
// }

// // Structs for the Request JSON objects
// type RequestClientSide struct {
// 	Op  string      `json:"op"`
// 	Key string      `json:"key"`
// 	Val interface{} `json:"val,omitempty"`
// }

// Sending a JSON-formatted HTTP request to the EMM server
func sendEMMRequest(op, key string, val interface{}) (interface{}, error) {
	// Conforming to the JSON request body format
	reqBody := Request {
		Key: key,
		Val: val,
		Op:  op,
	}
	
	// Marshalling the JSON request before it is sent to the EMM server
	buf, _ := json.Marshal(reqBody)
	resp, err := http.Post(endPoint, "application/json", bytes.NewReader(buf))
	if err != nil { // If we cannot send our request to the EMM server, we must error out
		return nil, err
	}
	defer resp.Body.Close()

	// Reading the raw HTTP body for debugging if the EMM server's status is not 2xx
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		rawBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(rawBody))
	}

	// Decoding the response from the EMM server
	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Result, nil

	// // Decoding the response from the EMM server
	// var result map[string]interface{}
	// if err := json.NewDecoder(resp.Body).Decode(&result); err != nil { // If the EMM server's response is wrong, we must error out
	// 	return nil, err
	// }
	
	// return result["result"], nil
}

// NOTE: The PUT/Update operation is unneeded because a working SET should always find the given index that it needs to PUT/Update
// Performing a legal GET/Read operation at the EMM server
func Get(key string) (interface{}, error) {
	return sendEMMRequest("get", key, nil)
}

// Performing a legal SET/Write operation at the EMM server
func Set(key string, val interface{}) error {
	_, err := sendEMMRequest("set", key, val)
	return err
}

// Performing a legal DEL/Remove operation at the EMM server
func Delete(key string) error {
	_, err := sendEMMRequest("del", key, nil)
	return err
}

