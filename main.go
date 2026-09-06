package main

import (
    "fmt"
    "os"
	"strings"
    "github.com/andelf/go-curl"
)

const (
	hideCursor = "\033[?25l"
	showCursor = "\033[?25h"
	clearLine  = "\033[2K"
)

var unit = []string{
	"Bytes", "KB", "MB", "GB", "TB", "PB",
}

type DL_DATA struct {
	URL		  string
	name	  string
	File	  *os.File
}

func writeCallback(data []byte, userdata interface{}) bool {
    f := userdata.(*os.File)
    f.Write(data)
    return true
}

func progressCallback(dltotal, dlnow, ultotal, ulnow float64, userdata interface{}) bool {
    if dltotal <= 0 {
		return true
	}
	size := float64(dlnow)
	total := float64(dltotal)

	ind_size := 0
	ind_total := 0

	for size >= 1024 && ind_size < 5 {
		size /=1024
		ind_size++

	}

	for total >= 1024 && ind_total < 5 {
		total /=1024
		ind_total++
	}


	percent := int((dlnow * 100) / dltotal)
	fmt.Printf("%s\r[%3d%%] %.2f %s / %.2f %s", clearLine, percent, size, unit[ind_size], total, unit[ind_total])

    return true
}

func extract_name(URL string) string {
	lastSlash := strings.LastIndex(URL, "/")
	if lastSlash != -1 && lastSlash < len(URL) -1 {
		filename := URL[lastSlash+1:]
		return filename
	}
	fmt.Printf("file not found")
	return "ERR"
}

func main() {
	// check args
	if len(os.Args) < 2 {
		fmt.Println("Usage: todlc [LINK]")
		return
	}

	// make download data context
	dctx := DL_DATA{ URL: os.Args[len(os.Args) - 1] }

	// extract name
	dctx.name = extract_name(dctx.URL)
	if dctx.name == "ERR" {
		fmt.Println("can't extract name!")
		os.Exit(-1)
	}

	// hide cursor
	fmt.Print(hideCursor) // hide cursor
	defer fmt.Print(showCursor) // show cursor in last

	// init curl
    curlru := curl.EasyInit()
    defer curlru.Cleanup()

	// init file
    f, err := os.Create(dctx.name)
    if err != nil {
        fmt.Println("Error", err)
        os.Exit(-1)
    }
	defer f.Close()


    // Use a large file to see progress
    curlru.Setopt(curl.OPT_URL, dctx.URL)
    curlru.Setopt(curl.OPT_WRITEFUNCTION, writeCallback)
    curlru.Setopt(curl.OPT_WRITEDATA, f)
    
    // Enable progress
    curlru.Setopt(curl.OPT_NOPROGRESS, false)
    curlru.Setopt(curl.OPT_PROGRESSFUNCTION, progressCallback)

    if err := curlru.Perform(); err != nil {
        fmt.Printf("\nError: %v\n", err)
    }

    fmt.Println("\nDone!")
}
