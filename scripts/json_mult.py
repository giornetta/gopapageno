#!/usr/bin/env python3

"""JSON Multiplier

This scripts reads a JSON file and writes a new JSON file obtained by replicated each key-value pair n times.
"""
import sys
import json
import os


def print_usage():
    print('Usage: json_mult source.json n')
    sys.exit(1)


if __name__ == '__main__':
    if len(sys.argv) < 3:
        print_usage()

    source_filename = sys.argv[1]
    num_copies = 1

    try:
        num_copies = int(sys.argv[2])
    except ValueError:
        print_usage()

    result = {}

    with open(source_filename) as f:
        data = json.load(f)

        if isinstance(data, dict):
            for i in range(num_copies):
                for k, v in data.items():
                    result[f"{k}_{i+1}"] = v
        else:
            raise TypeError(f"Unsupported JSON content type: {type(data)}")

    fn, ext = os.path.splitext(source_filename)
    dest_filename = f"{fn}-{num_copies}{ext}"
    with open(dest_filename, 'w') as f:
        json.dump(result, f, indent=2)
