#!/bin/bash

BENCH_NAME=BenchmarkParseOnly
COUNT=10

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -b|--bench)
            # Validate benchmark name
            if [[ "$2" != "BenchmarkParseOnly" && "$2" != "BenchmarkParse" ]]; then
                echo "Error: Benchmark name must be either 'BenchmarkParseOnly' or 'BenchmarkParse'"
                exit 1
            fi
            BENCH_NAME="$2"
            shift 2
            ;;
        -c|--count)
            # Validate count as a positive integer
            if ! [[ "$2" =~ ^[0-9]+$ ]] || [ "$2" -le 0 ]; then
                echo "Error: Count must be a positive integer"
                exit 1
            fi
            COUNT="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  -b, --bench BENCHMARK  Specify benchmark name (default: BenchmarkParseOnly)"
            echo "                         Allowed values: BenchmarkParseOnly, BenchmarkParse"
            echo "  -c, --count COUNT      Specify benchmark count (default: 10)"
            echo "  -h, --help             Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Create base filename using benchmark name and count
BASE_FILE="benchmark_json_${BENCH_NAME}_c${COUNT}"

echo "Running benchmarks for OPP with benchmark '$BENCH_NAME' and count $COUNT..."
go test github.com/giornetta/gopapageno/examples/json/opp -bench=$BENCH_NAME -count=$COUNT > ${BASE_FILE}_opp.txt
echo 'Running benchmarks for AOPP...'
go test github.com/giornetta/gopapageno/examples/json/aopp -bench=$BENCH_NAME -count=$COUNT > ${BASE_FILE}_aopp.txt
echo 'Running benchmarks for COPP...'
go test github.com/giornetta/gopapageno/examples/json/copp -bench=$BENCH_NAME -count=$COUNT > ${BASE_FILE}_copp.txt

echo "Generating figure ${BASE_FILE}.png"
python3 ./scripts/combine_benchmarks.py ${BASE_FILE}_opp.txt ${BASE_FILE}_aopp.txt ${BASE_FILE}_copp.txt > ${BASE_FILE}_combine.txt
go tool benchplot -bench=$BENCH_NAME -group-by=strategy -plots=avg_line -o=${BASE_FILE}.png -x=goroutines -top-legend ${BASE_FILE}_combine.txt
rm ${BASE_FILE}_combine.txt

echo "Generating benchstat file ${BASE_FILE}_compare.txt"
python3 prepare_benchmarks.py ${BASE_FILE}_opp.txt ${BASE_FILE}_aopp.txt ${BASE_FILE}_copp.txt ${BASE_FILE}_stat.txt
go tool benchstat ${BASE_FILE}_stat.txt > ${BASE_FILE}_compare.txt
rm ${BASE_FILE}_stat.txt

echo "Benchmark completed successfully!"
