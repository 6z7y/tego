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
    -v          show version
    -h          show this help
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
				fmt.Fprintln(os.Stderr, "-t required a num between range (1..32)")
				return false
			}
			n, err := strconv.Atoi((*args)[i+1])
			if err != nil || n < 1 || n > 32 {
				fmt.Fprintln(os.Stderr, "invalid thread count  (1..32")
				return false
			}
			(*dl).per_thread = n
			*args = append((*args)[:i], (*args)[i+2:]...)
		case "-v":
			fmt.Printf("%s: %s\n", NAME_PROG, VER_PROG)
			return true
		case "-h":
			fmt.Println(HELP_MSG)
			return true
		default:
			i++
		}
	}
	return true
}
