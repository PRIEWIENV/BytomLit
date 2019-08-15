#!/usr/bin/python3
# -*- coding: UTF-8 -*-

import sys, getopt, json, ast
import bytom_rpc

def main():
    in_file = ''
    try:
        opts, args = getopt.getopt(sys.argv[1:],"hi:")
    except getopt.GetoptError as err:
        print(err)
        print('submit_tx.py -i <inputfile>')
        sys.exit(2)
    for opt, arg in opts:
        if opt == '-h':
            print('submit_tx.py -i <inputfile>')
            sys.exit()
        elif opt == "-i":
            in_file = arg
        else:
            assert False, "unhandled option"

    if not in_file:
        print('submit_tx.py -i <inputfile>')
        sys.exit(2)
    with open(in_file, 'r') as f:
        data = json.loads(f.read())
    r = bytom_rpc.call('submit-transaction', data)
    r_str = json.dumps(r, sort_keys=True, indent=2, separators=(',', ':'))
    print(r_str)

if __name__ == "__main__":
   main()
