#!/usr/bin/env python3

"""Prepare Benchmarks

This scripts takes as input a number of Go benchmark files, possibly obtained by benchmarks in different packages,
and modifies them so that they can be used to compare them with utilities such as `benchstat`.
"""
import sys
import os
import subprocess


def print_usage():
    print('Usage: prepare_benchmarks file1 file2 ... fileN outfile')
    sys.exit(1)


if __name__ == '__main__':
    if len(sys.argv) < 4:
        print_usage()

    strategies = ['copp', 'aopp', 'opp']

    stat = []

    for i, filename in enumerate(sys.argv[1:-1]):
        with open(filename) as f:
            lines = f.read()

            for strat in strategies:
                lines = lines.replace(strat, '')
                lines = lines.replace(strat.upper(), '')

        fn, ext = os.path.splitext(filename)
        statfname = fn + "_stat" + ext
        with open(statfname, 'w') as f:
            f.write(lines)

        stat.append(statfname)

    with open(sys.argv[-1], 'w') as f:
        subprocess.run(['benchstat'] + stat, stdout=f)

    for fname in stat:
        os.remove(fname)