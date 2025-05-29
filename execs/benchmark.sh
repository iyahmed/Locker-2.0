#!/bin/bash

# Initalization settings


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

# Clearing the benchmark data file at the start
echo "Benchmark data log" > benchmark_data.txt


# Warm-up Phase: Mixing reads and writes
for ((i=1; i<=NUM_WARMUP_BATCHES; i++)); do
  echo -e "\n======= First Phase: Warm-up Request $i =======" >> benchmark_data.txt
  DATA="["
  for ((j=1; j<=BATCH_SIZE; j++)); do
        RAND=$((RANDOM % 100 + 1))
        if [ ${#WRITTEN_KEYS[@]} -gt 0 ] && [ $RAND -le $READ_PERCENTAGE ]; then
            OP="read"
            KEY=${WRITTEN_KEYS[$RANDOM % ${#WRITTEN_KEYS[@]}]}
        else
            OP="write"
            KEY=${KEYS[$RANDOM % ${#KEYS[@]}]}
            WRITTEN_KEYS+=("$KEY")
            # echo "DEBUG: Selected key for write: $KEY" # Debugging print statement
        fi
        USER=${USERS[$RANDOM % ${#USERS[@]}]}
        VAL=$(tr -dc 'a-zA-Z0-9' < /dev/urandom | head -c $MAX_VAL_SIZE | base64)
        ENTRY="{\"rid\":\"$USER\",\"op\":\"$OP\",\"key\":\"$KEY\",\"val\":\"\\\"$VAL\\\"\"}"
        DATA+="$ENTRY"
        if [ "$j" -lt "$BATCH_SIZE" ]; then
            DATA+=","
        fi
    done
    DATA+="]"
    echo "Data: $DATA" >> benchmark_data.txt
    RESPONSE=$(curl -s -X POST "$URL" -H "Content-Type: application/json" -d "$DATA")
    echo "Response: $RESPONSE" >> benchmark_data.txt
done


# Delete Phase: Deleting around half of the written keys
echo -e "\n======= Second Phase: Delete Request =======" >> benchmark_data.txt
DATA="["
DELETE_KEYS=("${WRITTEN_KEYS[@]:0:${#WRITTEN_KEYS[@]}/2}")
for i in "${!DELETE_KEYS[@]}"; do
    USER=${USERS[$RANDOM % ${#USERS[@]}]}
    ENTRY="{\"rid\":\"$USER\",\"op\":\"delete\",\"key\":\"${DELETE_KEYS[$i]}\",\"val\":\"\"}"
    DATA+="$ENTRY"
    if [ "$i" -lt "$((${#DELETE_KEYS[@]} - 1))" ]; then
        DATA+=","
    fi
done
DATA+="]"
echo "Data: $DATA" >> benchmark_data.txt
RESPONSE=$(curl -s -X POST "$URL" -H "Content-Type: application/json" -d "$DATA")
echo "Response: $RESPONSE" >> benchmark_data.txt


# Read Verification Phase: Verifying all the keys that could be read
echo -e "\n======= Third Phase: Final Read Verification =======" >> benchmark_data.txt
DATA="["
ALL_KEYS=($(printf "%s\n" "${WRITTEN_KEYS[@]}" | sort -u))
for i in "${!ALL_KEYS[@]}"; do
    USER=${USERS[$RANDOM % ${#USERS[@]}]}
    ENTRY="{\"rid\":\"$USER\",\"op\":\"read\",\"key\":\"${ALL_KEYS[$i]}\",\"val\":\"\"}"
    DATA+="$ENTRY"
    if [ "$i" -lt "$((${#ALL_KEYS[@]} - 1))" ]; then
        DATA+=","
    fi
done
DATA+="]"
echo "Data: $DATA" >> benchmark_data.txt
RESPONSE=$(curl -s -X POST "$URL" -H "Content-Type: application/json" -d "$DATA")
echo "Response: $RESPONSE" >> benchmark_data.txt
echo -e "\n======= Results: =======" >> benchmark_data.txt # Adding an extra line for visual seperation purposes


# Computing the execution/wall-clock time metric
end=$(date +%s.%N)
runtime=$(echo "$end - $start" | bc)

# Capturing the after benchmarking network and disk I/O
NET_IO_AFTER=$(cat /proc/net/dev | awk '/:/ {rx+=$2; tx+=$10} END{print rx, tx}')
DISK_IO_AFTER=$(cat /proc/diskstats | awk '{reads+=$6; writes+=$10} END{print reads, writes}')

# Compute the network and disk I/O metrics' differences and logging them
read_before=$(echo "$DISK_IO_BEFORE" | awk '{print $1}')
write_before=$(echo "$DISK_IO_BEFORE" | awk '{print $2}')
read_after=$(echo "$DISK_IO_AFTER" | awk '{print $1}')
write_after=$(echo "$DISK_IO_AFTER" | awk '{print $2}')
disk_read_diff=$((read_after - read_before))
disk_write_diff=$((write_after - write_before))
rx_before=$(echo "$NET_IO_BEFORE" | awk '{print $1}')
tx_before=$(echo "$NET_IO_BEFORE" | awk '{print $2}')
rx_after=$(echo "$NET_IO_AFTER" | awk '{print $1}')
tx_after=$(echo "$NET_IO_AFTER" | awk '{print $2}')
net_rx_diff=$((rx_after - rx_before))
net_tx_diff=$((tx_after - tx_before))
echo "Disk I/O Metrics: Reads=$disk_read_diff, Writes=$disk_write_diff sectors" | tee -a benchmark_data.txt
echo "Network I/O Metrics: Packets Received=$net_rx_diff bytes, Packets Sent=$net_tx_diff bytes" | tee -a benchmark_data.txt

# Logging the execution/wall-clock time metric
echo "The benchmark's runtime was $runtime seconds"
echo "Runtime: $runtime seconds" >> benchmark_data.txt
echo "Final configuration: warm-up batches: $NUM_WARMUP_BATCHES, batch size: $BATCH_SIZE, delete ops: ${#DELETE_KEYS[@]}, read verification: ${#ALL_KEYS[@]}" >> benchmark_data.txt

# Additional memory and CPU performance metrics
PID_CMD=$(ps -eo pid,comm --sort=-rss | head -n 2 | tail -n 1 | awk '{print $2}')
MAX_RSS=$(ps -eo rss,pid,comm --sort=-rss | head -n 2 | tail -n 1 | awk '{print $1}')
MEM_USAGE=$(free | grep Mem | awk '{printf("%.2f"), $3/$2 * 100.0}')
CPU_USAGE=$(top -b -n2 -d0.5 | grep "Cpu(s)" | tail -n1 | awk '{print $2 + $4}')

# Logging the additional memory and CPU peformance metrics
echo "Memory Usage: $MEM_USAGE%" | tee -a benchmark_data.txt
echo "CPU Usage: $CPU_USAGE%" | tee -a benchmark_data.txt
echo "Top Process RSS: $MAX_RSS KB" | tee -a benchmark_data.txt
