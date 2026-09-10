package main

import (
	"fmt"
	"os"
	"strconv"
	// "strings"
)

const HELP_MSG = `tego — multi-threaded download manager

Usage:
    tego <URL> [options]

Options:
    -t <num>    number of parallel threads (1..32, default 4)
	-o <file>	output filename (default: from url)
    -v          show version
    -h          show this help

Examples:
    tego https://6z7y.dpdns.org/randoms/tego_test.txt
    tego https://6z7y.dpdns.org/randoms/tego_test.txt -t 8
    tego https://6z7y.dpdns.org/randoms/tego_test.txt -t 8 -o myfile.txt
`

func arg_handle(args *[]string, dl *DL_DATA) bool {
	// args[0]: name program
	// args[1]: URL
	// args[2..]: options

	i := 1
	for i < len(*args) {
		switch (*args)[i] {
		case "-t":
			if i+1 >= len(*args) {
				fmt.Fprintln(os.Stderr, "-t requires a number (1..32)")
				return false
			}

			num, err := strconv.Atoi((*args)[i+1])
			if err != nil || num < 1 || num > 32 {
				fmt.Fprintln(os.Stderr, "thread count must be 1..32")
				return false
			}
			(*dl).per_thread = num
			*args = append((*args)[:i], (*args)[i+2:]...)

		case "-o":
			if i+1 >= len(*args) {
				fmt.Fprintln(os.Stderr, "-o requires a filename")
				return false
			}

			(*dl).name = (*args)[i+1]
			*args = append((*args)[:i], (*args)[i+2:]...)

		case "-v":
			fmt.Printf("%s: %s\n", NAME_PROG, VER_PROG)
			os.Exit(0)
		case "-h":
			fmt.Println(HELP_MSG)
			os.Exit(0)
		default:
			i++
		}
	}
	return true
}
