package main

import (
	"fmt"
	"os"
	"strings"
	"strconv"
	// "path/filepath"
	// "sync"
	// "time"
	"github.com/andelf/go-curl"
	"github.com/edsrzf/mmap-go"
)
type DL_CFG struct {
	per_thread  int
	// ststic path output
}

type Chunk struct {
	index		int
	start		int64
	end			int64
	downloaded	int64
	status		string
	// file		*os.File
	err			error
}

type DL_DATA struct {
	url			string
	name		string
	file		*os.File
	mmap		mmap.MMap
	size	    int64
	cfg			DL_CFG
	chunk  []*Chunk
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

// take size of file
func getSizeUrl(url string) (int64, error) {
	var size int64
	// init curl
	easy := curl.EasyInit()
	if easy == nil {
		return 0, fmt.Errorf("can't init curl")
		
	}

	defer easy.Cleanup()

	easy.Setopt(curl.OPT_URL, url) // url set
	easy.Setopt(curl.OPT_NOBODY, true) // without body information
	easy.Setopt(curl.OPT_FOLLOWLOCATION, true) // can follow source of link for real get data
	easy.Setopt(curl.OPT_WRITEFUNCTION, func(content []byte, userdata interface{}) bool { // discard the actual response body
		return true
	})

	easy.Setopt(curl.OPT_HEADERFUNCTION, func(data []byte, userdata interface{}) bool { // only take content-lenght value
		header := strings.TrimSpace(string(data))
		if strings.HasPrefix(strings.ToLower(header), "content-length:") {
			parts := strings.SplitN(header, ":", 2)
			if len(parts) == 2 {
				sizeStr := strings.TrimSpace(parts[1])
				if val, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
					if ptr, ok := userdata.(*int64); ok {
						*ptr = val
					}
				}
			}
		}
		return true
	})

	easy.Setopt(curl.OPT_WRITEHEADER, &size)

	err_easy := easy.Perform()
	if err_easy != nil {
		return 0, err_easy
	}

	return size, nil
}

func single_dl(dl *DL_DATA) error {
	// init a file
	f, err := os.Create(dl.name)
	if err != nil {
		return fmt.Errorf("can't create file %v", err)
	}
	dl.file = f
	defer dl.file.Close()

	// pre size empty blocks
	err = dl.file.Truncate(dl.size)
	if err != nil {
		return fmt.Errorf("can't pre allocate space: %v", err)
	}


	// link with mmap
	mmap, err := mmap.Map(dl.file, mmap.RDWR, 0)
	if err != nil {
		return fmt.Errorf("can't memory-map file: %v", err)
	}

	dl.mmap = mmap
	defer dl.mmap.Unmap()


	// init a curl
	easy := curl.EasyInit()
	if easy == nil {
		return fmt.Errorf("can't init curl")
	}

	defer easy.Cleanup()

	var offset int64 = 0

	easy.Setopt(curl.OPT_URL, dl.url) // link

	// write on a file using mmap
	easy.Setopt(curl.OPT_WRITEFUNCTION, func(data []byte, userdata interface{}) bool {
		dlPtr := userdata.(*DL_DATA)
		copy(dlPtr.mmap[offset:offset+int64(len(data))], data)
		offset += int64(len(data))
		return true
	})

	easy.Setopt(curl.OPT_WRITEDATA, dl)

	// progress
	easy.Setopt(curl.OPT_NOPROGRESS, false)
	easy.Setopt(curl.OPT_PROGRESSFUNCTION, progress_fn)

	err_easy := easy.Perform()
	if err_easy != nil {
		return fmt.Errorf("falied download: %v", err_easy)
	}
	return dl.mmap.Flush()
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

	// hide cursor
	fmt.Printf("%s", HIDE_CURSOR)
	defer fmt.Printf("%s", SHOW_CURSOR)

	// get size of a file
    size, err := getSizeUrl(os.Args[1])
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
	dl.size = size

	single_dl(&dl)
}
