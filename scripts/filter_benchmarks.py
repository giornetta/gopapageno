#!/usr/bin/env python3

"""Filter Benchmarks

This scripts takes as input a Go benchmark file, and filters lines which contain specified key value pairs.
"""
import sys
import os
import subprocess


def print_usage():
    print('Usage: filter_benchmarks file filters...')
    sys.exit(1)


if __name__ == '__main__':
    if len(sys.argv) < 3:
        print_usage()

    src_filename = sys.argv[1]
    filters = sys.argv[2:]

    with open(src_filename) as f:
        lines = f.read().splitlines()

        for line in lines:
            if ("Benchmark" not in line) or not any(fil not in line for fil in filters):
                print(line)
