#!/usr/bin/python3
# -*- coding: UTF-8 -*-

import sys, getopt, json
import bytom_rpc

def main():
    in_file = ''
    out_file = ''
    data = {}
    try:
        opts, args = getopt.getopt(sys.argv[1:],"hi:o:")
    except getopt.GetoptError as err:
        print(err)
        print('compile_contract.py -i <inputfile> -o <ouputfile>')
        sys.exit(2)
    if len(opts) != 2:
        print('compile_contract.py -i <inputfile> -o <outputfile>')
        sys.exit(2)
    for opt, arg in opts:
        if opt == '-h':
            print('compile_contract.py -i <inputfile> -o <outputfile>')
            sys.exit()
        elif opt in ("-i", "--ifile"):
            in_file = arg
        elif opt in ("-o", "--ofile"):
            out_file = arg
        else:
            assert False, "unhandled option"
    with open(in_file, 'r') as f:
        contract = f.read().replace('\n', '')
        data['contract'] = contract
    with open(out_file, 'w') as f:
        r = bytom_rpc.call('compile', data)
        r_str = json.dumps(r, sort_keys=True, indent=2, separators=(',', ':'))
        print(r_str)
        f.write(r_str)

if __name__ == "__main__":
   main()
