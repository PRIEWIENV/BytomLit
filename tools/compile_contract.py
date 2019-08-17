#!/usr/bin/python3
# -*- coding: UTF-8 -*-

import sys, getopt, json, ast
import bytom_rpc

def main():
    in_file = ''
    out_file = ''
    args_file = ''
    data = {}
    try:
        opts, args = getopt.getopt(sys.argv[1:],"hi:a:o:")
    except getopt.GetoptError as err:
        print(err)
        print('compile_contract.py -i <inputfile> [-a <argumentsfile>] -o <ouputfile>')
        sys.exit(2)
    for opt, arg in opts:
        if opt == '-h':
            print('compile_contract.py -i <inputfile> [-a <argumentsfile>] -o <ouputfile>')
            sys.exit()
        elif opt == "-i":
            in_file = arg
        elif opt == "-o":
            out_file = arg
        elif opt == "-a":
            args_file = arg
        else:
            assert False, "unhandled option"

    if not (in_file and out_file):
        print('compile_contract.py -i <inputfile> [-a <argumentsfile>] -o <ouputfile>')
        sys.exit(2)
    with open(in_file, 'r') as f:
        contract = f.read().replace('\n', '')
        data['contract'] = contract
        print(contract)
    if args_file:
        with open(args_file, 'r') as f:
            args = json.loads(f.read())
            data['args'] = args
    with open(out_file, 'w') as f:
        r = bytom_rpc.call('compile', data)
        r_str = json.dumps(r, sort_keys=True, indent=2, separators=(',', ':'))
        print(r_str)
        f.write(r_str)

if __name__ == "__main__":
   main()
