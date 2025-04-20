#!/bin/bash

echo 'Running benchmarks for OPP...'
go test github.com/giornetta/gopapageno/examples/json/opp -bench=BenchmarkParseOnly -count=10 > benchmark_json_opp.txt
echo 'Running benchmarks for AOPP...'
go test github.com/giornetta/gopapageno/examples/json/aopp -bench=BenchmarkParseOnly -count=10 > benchmark_json_aopp.txt
echo 'Running benchmarks for COPP...'
go test github.com/giornetta/gopapageno/examples/json/copp -bench=BenchmarkParseOnly -count=10 > benchmark_json_copp.txt

./scripts/combine_benchmarks.py benchmark_json_opp.txt benchmark_json_aopp.txt benchmark_json_copp.txt > benchmark_json_combine.txt
