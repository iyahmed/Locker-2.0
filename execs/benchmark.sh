#!/bin/bash

# Initalization settings


# The performance benchmarking requires `bc` to be installed locally
start=$(date +%s.%N)

# Default configuration parameters
NUM_REQUESTS=5                   # Number of requests to send
BATCH_SIZE=3                     # Size of a batch
MAX_VAL_SIZE=3                   # Maximum value size in bytes
READ_PERCENTAGE=50               # 50% reads and 50% writes by default
NUM_WARMUP_BATCHES=3             # Number of warm-up batches used
# URL="http://localhost:5000"    # The plaintext etcd client's URL
URL="http://localhost:5000/etcd" # The secure etcd client's URL

# User and key information
USERS=("user1" "user2" "user3" "user4" "user5" "user6" "user7" "user8" "user9" "user10")    # There are 10 users
DEFAULT_KEYS=("alpha" "beta" "gamma" "delta" "epsilon" "zeta" "eta" "theta" "iota" "kappa") # There are 10 default keys, with one key per user
WRITTEN_KEYS=()                                                                             # Storing all written keys for future reads

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
    -h|--help)
      echo "Usage: $0 [options]"
      echo "Options:"
      echo "  -n, --num-requests NUM      Number of requests to send (default: 5)"
      echo "  -b, --batch-size SIZE       Number of ops per request batch (default: 3)"
      echo "  -v, --val-size MAX          Maximum value size in bytes (default: 3)"
      echo "  -w, --warmup-batches NUM    Number of warm-up batches used (default: 5)"
      echo "  -r, --read-percentage PCT   Percentage of read operations (default: 50)"
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
            KEY=${DEFAULT_KEYS[$RANDOM % ${#DEFAULT_KEYS[@]}]}
            WRITTEN_KEYS+=("$KEY")
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

# Delete phase: Deleting around half of the written keys
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

# Final read-only phase: verify all keys
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


# OLD testing method that does not include phases and some edge cases:
#
# # Reading keys from file as defined by user input, each on a new line
# KEYS=()
# while IFS= read -r line; do
#     KEYS+=("$line")
# done < "large_keys.txt" # By default, all performance tests should be conducted on large_keys.txt for stress-testing purposes
# if [ "${#KEYS[@]}" -eq 0 ]; then
#   echo "Error: large_keys.txt is empty or missing"
#   exit 1
# fi
# 
# for ((i=1; i<=NUM_REQUESTS; i++)); do
#     # Generate the JSON array for the batch
#     DATA="["
    
#     for ((j=1; j<=BATCH_SIZE; j++)); do
#         # Determine operation based on read percentage
#         RAND=$((RANDOM % 100 + 1))
#         if [ ${#WRITTEN_KEYS[@]} -gt 0 ] && [ $RAND -le $READ_PERCENTAGE ]; then
#           OP="read"
#           KEY=${WRITTEN_KEYS[$RANDOM % ${#WRITTEN_KEYS[@]}]}
#         else
#           OP="write"
#           KEY=${KEYS[$RANDOM % ${#KEYS[@]}]}
#           WRITTEN_KEYS+=("$KEY")
#         fi
        
#         # Generate random user
#         USER=${USERS[$RANDOM % ${#USERS[@]}]}
        
#         # Generate random value (only needed for writes, but generate anyway)
#         VAL_SIZE=$((1 + RANDOM % MAX_VAL_SIZE))
#         VAL=$(tr -dc 'a-zA-Z0-9' < /dev/urandom | head -c $VAL_SIZE | base64)

#         # Add operation to batch
#         ENTRY="{\"rid\":\"$USER\",\"op\":\"$OP\",\"key\":\"$KEY\",\"val\":\"\\\"$VAL\\\"\"}"
#         DATA+="$ENTRY"
        
#         # Append comma if not last item
#         if [ "$j" -lt "$BATCH_SIZE" ]; then
#           DATA+=","
#         fi
#     done
    
#     DATA+="]"
    
#     # Log request information to benchmark_data.txt with append (>>)
#     echo -e "\n======= Request $i =======" >> benchmark_data.txt
#     echo "Data: $DATA" >> benchmark_data.txt
    
#     # Send curl request and capture response
#     RESPONSE=$(curl -s -X POST "$URL" -H "Content-Type: application/json" -d "$DATA")
    
#     # Log response
#     echo "Response: $RESPONSE" >> benchmark_data.txt
#     echo "Request $i sent with batch size: $BATCH_SIZE"
# done

# Performance benchmarking in seconds
end=$(date +%s.%N)
runtime=$(echo "$end - $start" | bc)
echo "The benchmark's runtime was $runtime seconds."
echo "Runtime: $runtime seconds" >> benchmark_data.txt
echo "Final configuration: warm-up batches: $NUM_WARMUP_BATCHES, batch size: $BATCH_SIZE, delete ops: ${#DELETE_KEYS[@]}, read verification: ${#ALL_KEYS[@]}" >> benchmark_data.txt
# OLD testing method: echo "Final configuration: $NUM_REQUESTS requests, $BATCH_SIZE batch size, $READ_PERCENTAGE% reads" >> benchmark_data.txt
