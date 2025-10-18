#!/bin/bash

if [ $# -ne 1 ]; then
    echo "Usage: $0 <directory without trailing slash>"
    exit 1
fi

if [ ! -d "$1" ]; then
    echo "Error: Directory $1 does not exist"
    exit 1
fi

output_file=$1'.csv'
echo "Filename,Result,Elapsed_time_ms,Config" > "$output_file"

for file in "$1"/*.cnf; do
    if [ -f "$file" ]; then
        output=$(go run satsolver -f "$file" -c mc 2>&1)

        filename=$(basename "$file")
        result=$(echo "$output" | grep "Result" | sed -E 's/.*Result: ([A-Z]+).*/\1/')
        elapsed_time=$(echo "$output" | grep "Elapsed time" | sed -E 's/.*Elapsed time: ([0-9]+\.[0-9]+) ms.*/\1/')
        config=$(echo "$output" | grep "Config" | sed -E 's/.*Config: ([a-zA-Z0-9-]+).*/\1/')        
        
        echo "$filename,$result,$elapsed_time,$config" >> "$output_file"
    fi
done
