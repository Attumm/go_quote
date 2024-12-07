#!/bin/bash

timestamp=$(date +"%Y_%m_%d_%H_%M")
start_seconds=$(date +%s)

echo "Starting benchmarks, please wait. This should take some time. ${timestamp}"

output_dir="benchmark/results"
mkdir -p "$output_dir"

bash scripts/benchmark/multi_benchmark.sh > "${output_dir}/${timestamp}_results.txt" 2>&1

bash scripts/benchmark/read_benchmark.sh < \
"${output_dir}/${timestamp}_results.txt" > "${output_dir}/${timestamp}_parsed.txt"

bash scripts/benchmark/make_benchmark_plot.sh < \
"${output_dir}/${timestamp}_parsed.txt" > "${output_dir}/${timestamp}_plots.txt"

elapsed_time=$(( $(date +%s) - start_seconds ))
echo "Benchmarks completed. Time taken: ${elapsed_time} seconds."

echo "Results stored ${output_dir}/${timestamp}_results.txt"
echo "Parsed result stored ${output_dir}/${timestamp}_parsed.txt"
echo "Plots stored ${output_dir}/${timestamp}_plots.txt"
