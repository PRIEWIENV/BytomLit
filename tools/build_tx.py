#!/usr/bin/python3
# -*- coding: UTF-8 -*-

import sys, getopt, json, ast
import bytom_rpc

def main():
    in_file = ''
    out_file = ''

    try:
        opts, args = getopt.getopt(sys.argv[1:],"hi:o:")
    except getopt.GetoptError as err:
        print(err)
        print('build_tx.py -i <inputfile> -o <ouputfile>')
        sys.exit(2)
    for opt, arg in opts:
        if opt == '-h':
            print('build_tx.py -i <inputfile> -o <ouputfile>')
            sys.exit()
        elif opt == "-i":
            in_file = arg
        elif opt == "-o"::
            out_file = arg
        else:
            assert False, "unhandled option"

    if not (in_file and out_file):
        print('build_tx.py -i <inputfile> -o <ouputfile>')
        sys.exit(2)
    with open(in_file, 'r') as f:
        data = json.loads(f.read())
    with open(out_file, 'w') as f:
        r = bytom_rpc.call('build-transaction', data)
        r_str = json.dumps(r, sort_keys=True, indent=2, separators=(',', ':'))
        print(r_str)
        f.write(r_str)

if __name__ == "__main__":
   main()
