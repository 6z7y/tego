package main

import (
	"fmt"
	"os"
	// "strings"
	"github.com/andelf/go-curl"
)
type DL_CFG struct {
	per_thread  int
	// ststic path output
}

type DL_DATA struct {
	url			string
	name		string
	file		*os.File
	file_err    error
	cfg			DL_CFG
}

var unit = [6]string{"B", "KB", "MB", "GB", "TB", "PB"}

func progress_fn(dltotal, dlnow, ultotal, ulnow float64, userdata interface{}) bool {
	if dltotal > 0 {
		percent := (dlnow / dltotal) * 100

		ind_now := 0
		ind_total := 0

		b_now := dlnow
		b_total := dltotal

		for b_now >= 1024 && ind_now < 5 {
			b_now /= 1024
			ind_now++;
		}

		for b_total >= 1024 && ind_total < 5 {
			b_total /= 1024
			ind_total++;
		}

		fmt.Printf("\r[%.2f%%] %.2f %s / %.2f %s         ", percent, b_now, unit[ind_now], b_total, unit[ind_total])

	}
	return true
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
		fmt.Errorf("can't extract name!")
		return
	}

	load_cfg(&dl.cfg)

	// init file
	dl.file, dl.file_err = os.Create(dl.name)
	if dl.file_err != nil {
		fmt.Errorf("can't create file %v", dl.file_err)
		return

	}
	defer dl.file.Close()

	// hide cursor
	fmt.Printf("%s", HIDE_CURSOR)
	defer fmt.Printf("%s", SHOW_CURSOR)

	// init curl
	easy := curl.EasyInit()
	if easy == nil {
		fmt.Errorf("can't init curl")
		return
	}

	defer easy.Cleanup()

	easy.Setopt(curl.OPT_URL, dl.url) // link

	// write on a file
	easy.Setopt(curl.OPT_WRITEDATA, dl.file)
	easy.Setopt(curl.OPT_WRITEFUNCTION, func(data []byte, userdata interface{}) bool {
		dl.file.Write(data)
		return true
	})

	// progress
	easy.Setopt(curl.OPT_NOPROGRESS, false)
	easy.Setopt(curl.OPT_PROGRESSFUNCTION, progress_fn)

	err_easy := easy.Perform()
	if err_easy != nil {
		fmt.Errorf("falied download: %v", err_easy)
		return
	}
}
