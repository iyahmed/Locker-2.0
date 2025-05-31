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
	"io/ioutil"
	"runtime"
	"sort"
	"strconv"
	"syscall"
	"path/filepath"
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
	tempDir string
}
type RealEMMAdapter struct {
	client *emm.Client
}
type MultiMap interface {
	Read(key string) [] string
	Write(key, val string)
	Delete(key string)
}
type DiskStats struct {
	ReadBytes       int64
	WriteBytes      int64
	CacheReadBytes  int64
	CacheWriteBytes int64
}

// Plaintext MultiMap implementation functions

// Constructing the Plaintext MultiMap
func newPlaintextMultiMap() *PlaintextMultiMap {
	// Creating a temporary file for the plaintext multimap's disk read/writes
	baseDir := "/mnt/data"
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		baseDir = os.TempDir() // fallback to /tmp or system default
	}
	dir, err := ioutil.TempDir(baseDir, "plaintext_multimap")
	if err != nil {
		log.Fatalf("Could not create a temporary directory: %v", err)
	}
	
	// Returning the Plaintext MultiMap Wrapper, with the storage as a map of arrays of strings
	return &PlaintextMultiMap {
		store: make(map[string][]string),
		tempDir: dir,
	}
}

// Reading the corresponding value, given a key, from the Plaintext MultiMap
func (p *PlaintextMultiMap) Read(key string) []string {
	// Mutex logic
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Returning the corresponding value paired to the key, if possible
	val := p.store[key]
	// Writing it into the temporary file
	path := filepath.Join(p.tempDir, fmt.Sprintf("read_%d_%s", time.Now().UnixNano(), key))
	f, _ := os.Create(path)
	pad := make([]byte, 4096)
	copy(pad, []byte(strings.Join(val, ",")))
	_, _ = f.Write(pad)
	_ = f.Sync()
	_ = f.Close()
	_, _ = ioutil.ReadFile(path)
	syscall.Sync()
	return val
}

// Writing the (key, value) pair into the Plaintext MultiMap
func (p *PlaintextMultiMap) Write(key, val string) {
	// Mutex logic
	p.mu.Lock()
	defer p.mu.Unlock()

	// Writing the (key, value) pair into the Plaintext MultiMap, if possible
	p.store[key] = append(p.store[key], val)
	// Writing it into a temporary file
	path := filepath.Join(p.tempDir, fmt.Sprintf("write_%d_%s", time.Now().UnixNano(), key))
	f, _ := os.Create(path)
	pad := make([]byte, 4096)
	copy(pad, []byte(val))
	_, _ = f.Write(pad)
	_ = f.Sync()
	_ = f.Close()
	syscall.Sync()
}

// Deleting the (key, value) pair from the Plaintext MultiMap
func (p *PlaintextMultiMap) Delete(key string) {
	// Mutex logic
	p.mu.Lock()
	defer p.mu.Unlock()

	// Deleting the (key, value) pair from the Plaintext MultiMap, if possible
	delete(p.store, key)
	// Writing it into a temporary file
	path := filepath.Join(p.tempDir, fmt.Sprintf("delete_%d_%s", time.Now().UnixNano(), key))
	f, _ := os.Create(path)
	pad := make([]byte, 4096)
	copy(pad, []byte(key))
	_, _ = f.Write(pad)
	_ = f.Sync()
	_ = f.Close()
	syscall.Sync()
}

// Baseline EMM Wrapper functions

// Constructing the Baseline EMM Wrapper
func newRealEMMWrapper() MultiMap {
	// Using the `EMM_insecure_client.go` file as the Baseline EMM's implementation
	client := emm.NewClient()

	// Returning the Baseline EMM Wrapper, as defined in the `emm` Go Module
	return &RealEMMAdapter{client: client}
}

// Reading the corresponding value, given a key, from the Baseline EMM
func (r *RealEMMAdapter) Read(key string) []string {
	// Reading the corresponding value, given a key, from the Baseline EMM
	vals, err := r.client.Get(key)
	if err != nil {
		log.Printf("Read error: %v", err)
		return nil
	}

	// Returning the corresponding value, given a key, from the Baseline EMM
	return vals
}

// Writing the (key, value) pair into the Baseline EMM
func (r *RealEMMAdapter) Write(key, val string) {
	// Writing the (key, value) pair into the Baseline EMM, if possible
	if err := r.client.Put(key, val); err != nil {
		log.Printf("Write error: %v", err)
	}
}

// Deleting the (key, value) pair from the Baseline EMM
func (r *RealEMMAdapter) Delete(key string) {
	// Deleting the (key, value) pair from the Baseline EMM, if possible
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

// Tracking the current state of the disk
func readDiskIO() DiskStats {
	syscall.Sync() // Forced flush to disk, to get accurate disk write numbers
	stats := DiskStats{}
	if data, err := ioutil.ReadFile("/proc/self/io"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "read_bytes:") {
				stats.ReadBytes, _ = strconv.ParseInt(strings.Fields(line)[1], 10, 64)
			} else if strings.HasPrefix(line, "write_bytes:") {
				stats.WriteBytes, _ = strconv.ParseInt(strings.Fields(line)[1], 10, 64)
			} else if strings.HasPrefix(line, "rchar:") {
				stats.CacheReadBytes, _ = strconv.ParseInt(strings.Fields(line)[1], 10, 64)
			} else if strings.HasPrefix(line, "wchar:") {
				stats.CacheWriteBytes, _ = strconv.ParseInt(strings.Fields(line)[1], 10, 64)
			}
		}
	}
	return stats
}

// Recording the benchmark's metrics, just like in `benchmark.sh` and `repeated_benchmarks.sh`
func getUsageStats(startTime time.Time, before DiskStats, after DiskStats) (memUsage float64, cpuUsage float64, topRSS int, diskReads int64, diskWrites int64, cacheReads int64, cacheWrites int64) {
	// Finding out the current process's memory consumption metric (in Kilobytes)
	mem := &runtime.MemStats{}
	runtime.ReadMemStats(mem)
	memUsage = float64(mem.Alloc) / float64(mem.Sys) * 100.0

	// Finding out the top process' Resident Set Size (RSS) metric (in Kilobytes)
	topRSS = 0
	if data, err := ioutil.ReadFile("/proc/self/status"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "VmRSS:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					topRSS, _ = strconv.Atoi(parts[1])
				}
				break
			}
		}
	}

	// Finding out the current process' CPU consumption metric (in percentages)
	var usage syscall.Rusage
	syscall.Getrusage(syscall.RUSAGE_SELF, &usage)
	cpuUser := time.Duration(usage.Utime.Sec) * time.Second + time.Duration(usage.Utime.Usec) * time.Microsecond
	cpuSys := time.Duration(usage.Stime.Sec) * time.Second + time.Duration(usage.Stime.Usec) * time.Microsecond
	totalCPU := cpuUser + cpuSys
	totalElapsed := time.Since(startTime)
	cpuCount := float64(runtime.NumCPU())
	cpuUsage = (totalCPU.Seconds() / totalElapsed.Seconds()) * 100.0 / cpuCount

	// Finding out the current process' disk utilization metrics (in sectors)
	diskReads = (after.ReadBytes - before.ReadBytes) / 512
	diskWrites = (after.WriteBytes - before.WriteBytes) / 512
	cacheReads = (after.CacheReadBytes - before.CacheReadBytes) / 512
	cacheWrites = (after.CacheWriteBytes - before.CacheWriteBytes) / 512

	// Returning the 7 metrics explicitly
	return memUsage, cpuUsage, topRSS, diskReads, diskWrites, cacheReads, cacheWrites
}

// Testing and benchmarking the Baseline EMM and the Plaintext MultiMap, using the same benchmarking methods as in `repeated_benchmarks.sh`
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

	// List of each individual runs for each metric
	runtimes := []float64{}
	memUsages := []float64{}
	cpuUsages := []float64{}
	rssValues := []float64{}
	diskReads := []float64{}
	diskWrites := []float64{}
	cacheReads := []float64{}
	cacheWrites := []float64{}

	for run := 1; run <= totalRuns; run++ { // Looping for each individual run
		// Initlization for an individual run
		diskBefore := readDiskIO()
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
		elapsed := float64(time.Since(start).Microseconds()) / 1000.0
		runtimes = append(runtimes, elapsed)
		diskAfter := readDiskIO()
		mem, cpu, rss, dreads, dwrites, creads, cwrites := getUsageStats(start, diskBefore, diskAfter)
		memUsages = append(memUsages, mem)
		cpuUsages = append(cpuUsages, cpu)
		rssValues = append(rssValues, float64(rss))
		diskReads = append(diskReads, float64(dreads))
		diskWrites = append(diskWrites, float64(dwrites))
		cacheReads = append(cacheReads, float64(creads))
		cacheWrites = append(cacheWrites, float64(cwrites))

		fmt.Printf("%s Run %d completed in %.3f seconds | Memory: %.2f%% | CPU: %.2f%% | Top-Process's RSS: %d KB | Disk Reads: %d sectors | Disk Writes: %d sectors | Disk Cache Reads: %d sectors | Disk Cache Writes: %d sectors \n",
			name, run, elapsed, mem, cpu, rss, dreads, dwrites, creads, cwrites)
	}

	// Outputting the average runs' (Middle 3 of 7 Runs) metrics
	avgRuntime := averageFloat64(runtimes)
	avgMem := averageFloat64(memUsages)
	avgCPU := averageFloat64(cpuUsages)
	avgRSS := averageFloat64(rssValues)
	avgDReads := averageFloat64(diskReads)
	avgDWrites := averageFloat64(diskWrites)
	avgCReads := averageFloat64(cacheReads)
	avgCWrites := averageFloat64(cacheWrites)
	fmt.Printf("\n\nAverage Metrics for %s (Middle 3 of 7 Runs):\n", name)
	fmt.Printf("Execution-Time/Wall-Clock Time/Runtime: %.3f Seconds\n", avgRuntime)
	fmt.Printf("Memory Utilization: %.2f%%\n", avgMem)
	fmt.Printf("CPU Utilization: %.2f%%\n", avgCPU)
	fmt.Printf("Top Process' Resident Set Size (RSS): %.0f KB\n", avgRSS)
	fmt.Printf("Direct Disk Reads: %.0f sectors\n", avgDReads)
	fmt.Printf("Direct Disk Writes: %.0f sectors\n", avgDWrites)
	fmt.Printf("Indirect Cache Reads: %.0f sectors\n", avgCReads)
	fmt.Printf("Indirect Cache Writes: %.0f sectors\n\n", avgCWrites)
	// return avg
	return 0
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
	benchmark("Baseline Secure EMM", realEMM, keys, users)
}
