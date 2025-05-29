package main


import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
	"math/rand"
	"os"
	"encoding/json"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"locker/libs/emm"
)


// Helper structs and interfaces
type Request struct {
	RID string          `json:"rid"`
	Op  string          `json:"op"`
	Key string          `json:"key"`
	Val json.RawMessage `json:"val"`
}
type PlaintextMultiMap struct {
	mu sync.RWMutex
	store map[string][]string
}
type RealEMMAdapter struct {
	client *emm.Client
}
type MultiMap interface {
	Read(key string) [] string
	Write(key, val string)
	Delete(key string)
}

// Plaintext MultiMap implementation functions

// Constructing the Plaintext MultiMap
func newPlaintextMultiMap() *PlaintextMultiMap {
	// Returning the Plaintext MultiMap Wrapper, with the storage as a map of arrays of strings
	return &PlaintextMultiMap {
		store: make(map[string][]string),
	}
}

// Reading the corresponding value, given a key, from the Plaintext MultiMap
func (p *PlaintextMultiMap) Read(key string) []string {
	// Mutex logic
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Returning the corresponding value paired to the key, if possible
	return p.store[key]
}

// Writing the (key, value) pair into the Plaintext MultiMap
func (p *PlaintextMultiMap) Write(key, val string) {
	// Mutex logic
	p.mu.Lock()
	defer p.mu.Unlock()

	// Writing the (key, value) pair into the Plaintext MultiMap, if possible
	p.store[key] = append(p.store[key], val)
}

// Deleting the (key, value) pair from the Plaintext MultiMap
func (p *PlaintextMultiMap) Delete(key string) {
	// Mutex logic
	p.mu.Lock()
	defer p.mu.Unlock()

	// Deleting the (key, value) pair from the Plaintext MultiMap, if possible
	delete(p.store, key)
}

// Real EMM Wrapper functions

// Constructing the Real EMM Wrapper
func newRealEMMWrapper() MultiMap {
	// Using the `EMM_insecure_client.go` file as the Real EMM's implementation
	client := emm.NewClient()

	// Returning the Real EMM Wrapper, as defined in the `emm` Go Module
	return &RealEMMAdapter{client: client}
}

// Reading the corresponding value, given a key, from the Real EMM
func (r *RealEMMAdapter) Read(key string) []string {
	// Reading the corresponding value, given a key, from the Real EMM
	vals, err := r.client.Get(key)
	if err != nil {
		log.Printf("Read error: %v", err)
		return nil
	}
	return vals
}

// Writing the (key, value) pair into the Real EMM
func (r *RealEMMAdapter) Write(key, val string) {
	// Writing the (key, value) pair into the Real EMM, if possible
	if err := r.client.Put(key, val); err != nil {
		log.Printf("Write error: %v", err)
	}
}

// Deleting the (key, value) pair from the Real EMM
func (r *RealEMMAdapter) Delete(key string) {
	// Deleting the (key, value) pair from the Real EMM, if possible
	if err := r.client.Delete(key); err != nil {
		log.Printf("Delete error: %v", err)
	}
}

// Utility functions to ease with the testing and benchmarking

// Computing a random English-alphabet alphanumeric value
func randomValue(size int) string {
	// Creating an alphanumeric code
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, size)
	// Selecting a random number from the alphanumeric code
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	// Returning the randomized alphanumeric value as a string
	return string(b)
}

// Computing the mean of an array of 64-bit floating-point numbers, given 5 entries
func averageFloat64(values []float64) float64 {
	// Sorting values by ascending order
	sort.Float64s(values)
	// Ensuring that we have enough data for data accuracy purposes
	if len(values) < 5 {
		return 0
	}
	// Filtering out the 4 most extreme values (2 largest and 2 smallest)
	values = values[2 : len(values) - 2]
	// Mean formula
	sum := 0.0
	for _, v := range values {
		sum += v
	}

	// Returning the filtered average of at least 5 64-bit floating-point numbers
	return sum / float64(len(values))
}

// Reading the key file from the operating system (OS)
func readKeys(filename string) []string {
	// Reading the given filename, if possible
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read key file: %v", err)
	}
	// Iterating over each line and trimming all space/newline characters
	lines := strings.Split(string(content), "\n")
	var keys []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			// Appending the key(s) in each line into the return array
			keys = append(keys, trimmed)
		}
	}

	// Returning all the keys found in the given key file as an array
	return keys
}

// Recording the benchmark's metrics, just like in `benchmark.sh` and `repeated_benchmarks.sh`
func getUsageStats() (memUsage float64, cpuUsage float64, topRSS int, diskReads int, diskWrites int) {
	// Finding out the current process's memory consumption metric (in Kilobytes)
	mem := &runtime.MemStats{}
	runtime.ReadMemStats(mem)
	memUsage = float64(mem.Alloc) / float64(mem.Sys) * 100.0

	// Finding out the top process' Resident Set Size (RSS) metric (in Kilobytes)
	topRSS = 0
	topCmd := exec.Command("ps", "-eo", "rss", "--sort=-rss")
	output, err := topCmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		if len(lines) > 1 {
			topRSS, _ = strconv.Atoi(strings.TrimSpace(lines[1]))
		}
	}

	// Finding out the current process' CPU consumption metric (in percentages)
	cpuCmd := exec.Command("top", "-bn2", "-d0.5")
	cpuOut, err := cpuCmd.Output()
	if err == nil {
		lines := strings.Split(string(cpuOut), "\n")
		for _, line := range lines {
			if strings.Contains(line, "CPU(s)") {
				fields := strings.Fields(line)
				if len(fields) >= 8 {
					u, _ := strconv.ParseFloat(strings.Trim(fields[1], "%"), 64)
					s, _ := strconv.ParseFloat(strings.Trim(fields[3], "%"), 64)
					cpuUsage = u + s
				}
			}
		}
	}

	// Finding out the current process' disk utlization metrics (in sectors)
	diskCmd := exec.Command("awk", "{reads+=$6; writes+=$10} END{print reads, writes}", "/proc/diskstats")
	diskOut, err := diskCmd.Output()
	if err == nil {
		parts := strings.Fields(string(diskOut))
		if len(parts) == 2 {
			diskReads, _ = strconv.Atoi(parts[0])
			diskWrites, _ = strconv.Atoi(parts[1])
		}
	}

	// Void return, to avoid Go compiler errors
	return
}

// Testing and benchmarking the Real EMM and the Plaintext MultiMap, using the same benchmarking methods as in `repeated_benchmarks.sh`
func benchmark(name string, mmap MultiMap, keys, users []string) float64 {
	// These are the same default setttings as in `repeated_benchmarks.sh`
	const (
		numRequests      = 5
		batchSize        = 3
		valSize          = 3
		numWarmupBatches = 3
		readPercentage   = 50
		totalRuns        = 7
	)

	// Looping for each individual run
	runtimes := []float64{}
	for run := 1; run <= totalRuns; run++ {
		// Initlization for an individual run
		start := time.Now()
		writtenKeys := []string{}
		var allKeys []string

		// Phase 1: Warm-Up Phase with Mixed Reads and Writes)
		for i := 0; i < numWarmupBatches; i++ {
			for j := 0; j < batchSize; j++ {
				key := keys[rand.Intn(len(keys))]
				val := randomValue(valSize)
				op := "write"
				if len(writtenKeys) > 0 && rand.Intn(100) < readPercentage {
					op = "read"
					key = writtenKeys[rand.Intn(len(writtenKeys))]
				}
				user := users[rand.Intn(len(users))]
				switch op {
				case "write":
					mmap.Write(key, val)
					writtenKeys = append(writtenKeys, key)
				case "read":
					_ = mmap.Read(key)
				}
				_ = user
			}
		}

		// Phase 2: Delete Phase with the Deletiion of Half of the Already-Written Keys
		for i := 0; i < len(writtenKeys)/2; i++ {
			mmap.Delete(writtenKeys[i])
		}

		// Phase 3: Read Verification Phase with the Reading of the Remaning Non-Deleted Keys
		keySeen := map[string]bool{}
		for _, key := range writtenKeys {
			if !keySeen[key] {
				_ = mmap.Read(key)
				keySeen[key] = true
				allKeys = append(allKeys, key)
			}
		}

		// Computing the metrics for an individual run
		elapsed := time.Since(start).Seconds()
		runtimes = append(runtimes, elapsed)
		mem, cpu, rss, dreads, dwrites := getUsageStats()
		fmt.Printf("%s Run %d completed in %.3f seconds | Memory: %.2f%% | CPU: %.2f%% | Top-Process's RSS: %d KB | Disk Reads: %d sectors | Disk Writes: %d sectors\n",
			name, run, elapsed, mem, cpu, rss, dreads, dwrites)
	}

	// Outputting the average runs' metrics
	avg := averageFloat64(runtimes)
	fmt.Printf("\nAverage Runtime for %s (Middle 3 of 7 Runs): %.3f Seconds\n\n", name, avg)
	return avg
}


// Comparing the EMM implementation with the Plaintext MultiMap implementation, similar to `repeated_benchmarks.sh`
func main() {
	// Initalization Logic
	rand.Seed(time.Now().UnixNano())
	users := []string{"user1", "user2", "user3", "user4", "user5", "user6", "user7", "user8", "user9", "user10"}
	keys := readKeys("medium_keys.txt")
	plainTest := newPlaintextMultiMap()
	realEMM := newRealEMMWrapper()

	// Benchmarking logic
	benchmark("Plaintext MultiMap", plainTest, keys, users)
	benchmark("Real Secure EMM", realEMM, keys, users)
}
