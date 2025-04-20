#!/usr/bin/env python3

"""Combine Benchmarks

This scripts takes as input a number of Go benchmark files, possibly obtained by benchmarks in different packages,
and outputs a single file that can be used with utilities such as `benchplot`.
"""
import sys


def print_usage():
    print('Usage: combine_benchmarks file1 file2 ... fileN (at least two files are required)')
    sys.exit(1)


if __name__ == '__main__':
    if len(sys.argv) < 3:
        print_usage()

    last = []

    for i, filename in enumerate(sys.argv[1:]):
        with open(filename) as f:
            lines = f.read().splitlines()

            for j, line in enumerate(lines):
                if j < 4:
                    if i == 0:
                        print(line)
                elif j >= len(lines) - 2:
                    if i == 0:
                        last.append(line)
                else:
                    print(line)

    for line in last:
        print(line)
