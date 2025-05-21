#!/bin/bash

# The performance benchmarking requires `bc` to be installed locally
start=$(date +%s.%N)

# Default configuration parameters
NUM_REQUESTS=5         # Number of requests to send
BATCH_SIZE=3           # Size of a batch
MAX_VAL_SIZE=3         # Maximum value size in bytes
READ_PERCENTAGE=50     # 50% reads and 50% writes by default

# Process command line arguments
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

# Validate read percentage
if [ "$READ_PERCENTAGE" -lt 0 ] || [ "$READ_PERCENTAGE" -gt 100 ]; then
  echo "Error: Read percentage must be between 0 and 100"
  exit 1
fi

# Read keys from file, each on a new line
KEYS=()
while IFS= read -r line; do
    KEYS+=("$line")
done < "large_keys.txt"
if [ "${#KEYS[@]}" -eq 0 ]; then
  echo "Error: large_keys.txt is empty or missing"
  exit 1
fi

# There are 10 users
USERS=("user1" "user2" "user3" "user4" "user5" "user6" "user7" "user8" "user9" "user10")

# Store all written keys for future reads
WRITTEN_KEYS=()

# Target URL
# URL="http://localhost:5000"
URL="http://localhost:5000/etcd"

# Clear data file at the start
echo "Benchmark data log" > benchmark_data.txt

for ((i=1; i<=NUM_REQUESTS; i++)); do
    # Generate the JSON array for the batch
    DATA="["
    
    for ((j=1; j<=BATCH_SIZE; j++)); do
        # Determine operation based on read percentage
        RAND=$((RANDOM % 100 + 1))
        if [ ${#WRITTEN_KEYS[@]} -gt 0 ] && [ $RAND -le $READ_PERCENTAGE ]; then
          OP="read"
          KEY=${WRITTEN_KEYS[$RANDOM % ${#WRITTEN_KEYS[@]}]}
        else
          OP="write"
          KEY=${KEYS[$RANDOM % ${#KEYS[@]}]}
          WRITTEN_KEYS+=("$KEY")
        fi
        
        # Generate random user
        USER=${USERS[$RANDOM % ${#USERS[@]}]}
        
        # Generate random value (only needed for writes, but generate anyway)
        VAL_SIZE=$((1 + RANDOM % MAX_VAL_SIZE))
        VAL=$(tr -dc 'a-zA-Z0-9' < /dev/urandom | head -c $VAL_SIZE | base64)

        # Add operation to batch
        ENTRY="{\"rid\":\"$USER\",\"op\":\"$OP\",\"key\":\"$KEY\",\"val\":\"\\\"$VAL\\\"\"}"
        DATA+="$ENTRY"
        
        # Append comma if not last item
        if [ "$j" -lt "$BATCH_SIZE" ]; then
          DATA+=","
        fi
    done
    
    DATA+="]"
    
    # Log request information to benchmark_data.txt with append (>>)
    echo -e "\n======= Request $i =======" >> benchmark_data.txt
    echo "Data: $DATA" >> benchmark_data.txt
    
    # Send curl request and capture response
    RESPONSE=$(curl -s -X POST "$URL" -H "Content-Type: application/json" -d "$DATA")
    
    # Log response
    echo "Response: $RESPONSE" >> benchmark_data.txt
    echo "Request $i sent with batch size: $BATCH_SIZE"
done

# Performance benchmarking in seconds
end=$(date +%s.%N)
runtime=$(echo "$end - $start" | bc)

echo "The benchmark's runtime was $runtime seconds."
echo "Runtime: $runtime seconds" >> benchmark_data.txt
echo "Final configuration: $NUM_REQUESTS requests, $BATCH_SIZE batch size, $READ_PERCENTAGE% reads" >> benchmark_data.txt
