#!/bin/bash

# Initalization settings, same as `benchmark.sh`


# The performance benchmarking requires `bc` to be installed locally
start=$(date +%s.%N)

# Capturing the before-benchmarking network and disk I/O
NET_IO_BEFORE=$(cat /proc/net/dev | awk '/:/ {rx+=$2; tx+=$10} END{print rx, tx}')
DISK_IO_BEFORE=$(cat /proc/diskstats | awk '{reads+=$6; writes+=$10} END{print reads, writes}')

# Default configuration parameters
NUM_REQUESTS=5                   # Number of requests to send
BATCH_SIZE=3                     # Size of a batch
MAX_VAL_SIZE=3                   # Maximum value size in bytes
READ_PERCENTAGE=50               # 50% reads and 50% writes by default
NUM_WARMUP_BATCHES=3             # Number of warm-up batches used
KEY_FILE="large_keys.txt"        # Default file where keys are stored
SCRIPT_PATH="bash benchmark.sh"  # Location of the `benchmark.sh` file
# URL="http://localhost:5000"    # The proxy and plaintext etcd clients' URL
URL="http://localhost:5000/etcd" # The secure etcd client's URL

# User and key information
USERS=("user1" "user2" "user3" "user4" "user5" "user6" "user7" "user8" "user9" "user10")    # There are 10 users
WRITTEN_KEYS=()                                                                             # Storing all written keys for future reads
# Getting the keys from either the given key file or by some hardcoded back-up keys
KEYS=()
# Attempting to read from the given key file, if it is there
if [[ -f "$KEY_FILE" ]]; then
  while IFS= read -r line; do
    [[ -n "$line" ]] && KEYS+=("$line")
  done < "$KEY_FILE"
else
  echo "ERROR: The given key file '$KEY_FILE' is not found."
  exit 1
fi
# Checking if we can read from the given key file
if [ "${#KEYS[@]}" -eq 0 ]; then
  echo "ERROR: There are no valid keys found in $KEY_FILE"
  exit 1
fi

# Processing command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -n|--num-requests)
      NUM_REQUESTS="$2"
      shift 2
      ;;
    -b|--batch-size)
      BATCH_SIZE="$2"
      shift 2
      ;;
    -v|--val-size)
      MAX_VAL_SIZE="$2"
      shift 2
      ;;
    -w|--warmup-batches)
      NUM_WARMUP_BATCHES="$2"
      shift 2
      ;;
    -r|--read-percentage)
      READ_PERCENTAGE="$2"
      shift 2
      ;;
    -k|--key-file)
      KEY_FILE="$2"
      shift 2
      ;;
    -s|--script-path)
      SCRIPT_PATH="$2"
      shift 2
      ;;
    -h|--help)
      echo "Usage: $0 [options]"
      echo "Options:"
      echo "  -n, --num-requests NUM      Number of requests to send (default: 5)"
      echo "  -b, --batch-size SIZE       Number of ops per request batch (default: 3)"
      echo "  -v, --val-size MAX          Maximum value size in bytes (default: 3)"
      echo "  -w, --warmup-batches NUM    Number of warm-up batches used (default: 5)"
      echo "  -r, --read-percentage PCT   Percentage of read operations (default: 50)"
      echo "  -k, --key-file FILE         Path to the file where the keys are stored (default: large_keys.txt in the same folder as this benchmark.sh)"
      echo "  -h, --help                  Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use --help for usage information"
      exit 1
      ;;
  esac
done

# Validating the given read percentage
if [ "$READ_PERCENTAGE" -lt 0 ] || [ "$READ_PERCENTAGE" -gt 100 ]; then
  echo "Error: Read percentage must be between 0 and 100"
  exit 1
fi


# Creating arrays to store each `benchmark.sh` metric from each test run
RUNTIMES=()
MEM_USAGES=()
CPU_USAGES=()
RSS_VALUES=()
DISK_READS=()
DISK_WRITES=()
NET_RXS=()
NET_TXS=()

# Clearing the log file
echo "Benchmark data log for 7-run filtered average benchmark" > benchmark_data.txt


# Running the individual benchmark 7 times in a temporary file
for i in {1..7}; do
  # Initalizing the individual runs with the output going to a temporary file
  echo -e "\n==== Run $i ====" | tee -a benchmark_data.txt
  TMP_FILE=$(mktemp)
  $SCRIPT_PATH -n "$NUM_REQUESTS" -b "$BATCH_SIZE" -v "$MAX_VAL_SIZE" -w "$NUM_WARMUP_BATCHES" -r "$READ_PERCENTAGE" -k "$KEY_FILE" > "$TMP_FILE" 2>&1
  cat "$TMP_FILE" | tee -a benchmark_data.txt

  # Gathering the data from `benchmark.sh`
  while read -r line; do
    case "$line" in
      *runtime\ was*)
        RUNTIMES+=( $(echo "$line" | grep -oP '[0-9]+\.[0-9]+') )
        ;;
      *Memory\ Usage:*)
        MEM_USAGES+=( $(echo "$line" | grep -oP 'Memory Usage:\s*\K[0-9.]+') )
        ;;
      *CPU\ Usage:*)
        CPU_USAGES+=( $(echo "$line" | grep -oP 'CPU Usage:\s*\K[0-9.]+') )
        ;;
      *Top\ Process\ RSS:*)
        RSS_VALUES+=( $(echo "$line" | grep -oP 'Top Process RSS:\s*\K[0-9]+') )
        ;;
      *Disk\ I/O\ Metrics:*)
        DISK_READS+=( $(echo "$line" | awk -F'[=,]' '{print $2}' | awk '{print $1}') )
        DISK_WRITES+=( $(echo "$line" | awk -F'[=,]' '{print $4}' | awk '{print $1}') )
        ;;
      *Network\ I/O\ Metrics:*)
        NET_RXS+=( $(echo "$line" | awk -F'[=,]' '{print $2}' | awk '{print $1}') )
        NET_TXS+=( $(echo "$line" | awk -F'[=,]' '{print $4}' | awk '{print $1}') )
        ;;
    esac
  done < "$TMP_FILE"
  rm "$TMP_FILE"
done


# Trimming out the 4 most extreme test runs (2 largest and 2 smallest)
function trim_extremes {
  local array=("${!1}")
  local sorted=($(printf '%s\n' "${array[@]}" | sort -n))
  echo "${sorted[@]:2:3}"
}
TRIM_RUNTIMES=($(trim_extremes RUNTIMES[@]))
TRIM_MEM=($(trim_extremes MEM_USAGES[@]))
TRIM_CPU=($(trim_extremes CPU_USAGES[@]))
TRIM_RSS=($(trim_extremes RSS_VALUES[@]))
TRIM_DISK_READS=($(trim_extremes DISK_READS[@]))
TRIM_DISK_WRITES=($(trim_extremes DISK_WRITES[@]))
TRIM_NET_RXS=($(trim_extremes NET_RXS[@]))
TRIM_NET_TXS=($(trim_extremes NET_TXS[@]))

# Averaging out the remaining 3 test runs, while ensuring that we don't divide by 0
function average {
  local values=("$@")
  local total=0
  local count=${#values[@]}
  if [ "$count" -eq 0 ]; then
    echo "0"
    return
  fi
  for val in "${values[@]}"; do
    total=$(echo "$total + $val" | bc)
  done
  echo "scale=3; $total / $count" | bc
}
AVG_RUNTIME=$(average "${TRIM_RUNTIMES[@]}")
AVG_MEM=$(average "${TRIM_MEM[@]}")
AVG_CPU=$(average "${TRIM_CPU[@]}")
AVG_RSS=$(average "${TRIM_RSS[@]}")
AVG_DREAD=$(average "${TRIM_DISK_READS[@]}")
AVG_DWRITE=$(average "${TRIM_DISK_WRITES[@]}")
AVG_RX=$(average "${TRIM_NET_RXS[@]}")
AVG_TX=$(average "${TRIM_NET_TXS[@]}")


# Outputting and logging the final averaged metrics
echo -e "\n===== Final Averaged Metrics (Middle 3 of 7 Test Runs) ====="
{
  echo "Average Runtime: $AVG_RUNTIME seconds"
  echo "Average Memory Usage: $AVG_MEM%"
  echo "Average CPU Usage: $AVG_CPU%"
  echo "Average Peak RSS: $AVG_RSS KB"
  echo "Average Disk Reads: $AVG_DREAD sectors"
  echo "Average Disk Writes: $AVG_DWRITE sectors"
  echo "Average Network RX: $AVG_RX bytes"
  echo "Average Network TX: $AVG_TX bytes"
} | tee -a benchmark_data.txt
