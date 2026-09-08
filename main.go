package main

import (
	"fmt"
	"os"
	// "strings"
	// "github.com/andelf/go-curl"
)
type DL_CFG struct {
	per_thread  int
	// ststic path output
}

type DL_DATA struct {
	url			string
	name		string
	file		*os.File
	cfg			DL_CFG
}

func main() {
	// 1. arg checker
	if len(os.Args) < 2 {
		fmt.Printf("Usage: tego [LINK]")
		return
	}

	// 2. init context
	dl := DL_DATA{ url: os.Args[1] }

	// 3. extract name from url
	dl.name = extract_name(dl.url)
	if dl.name == "ERR" {
		fmt.Println("can't extract name!")
		return
	}

	load_cfg(&dl.cfg)
}
