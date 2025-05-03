// libs/hirb/HIRB.go
// Ismail Ahmed: Implements the HIRB oblivious memory data structure

package hirb

// Needed imports
import (
	"bytes"
	"encoding/json"
	// "fmt"
	"net/http"
)

// Listening for the internal HIRB HTTP server at Port 8236
const apiEndpoint = "http://localhost:8236"

// All HIRB requests will be handled here
func sendHIRBRequest(op, key string, val interface{}) (interface{}, error) {
	// fmt.Printf("Hey everyone!")
	// Request body handling
	reqBody := map[string]interface{}{
		"op":  op,
		"key": key,
	}
	if val != nil {
		reqBody["val"] = val
	}

	// Data marshalling and error handling
	buf, _ := json.Marshal(reqBody)
	resp, err := http.Post(apiEndpoint, "application/json", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Returning the results if it is valid
	return result["result"], nil
}


// GET/READ JSON operation
func Get(key string) (interface{}, error) {
	return sendHIRBRequest("get", key, nil)
}

// SET/WRITE JSON operation
func Set(key string, val interface{}) error {
	_, err := sendHIRBRequest("set", key, val)
	return err
}

// DELETE/REMOVE operation
func Delete(key string) error {
	_, err := sendHIRBRequest("del", key, nil)
	return err
}
