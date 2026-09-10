package main

import (
	"fmt"
	"os"
	// "strings"
	// "strconv"
	"sync"
	"time"
	"github.com/andelf/go-curl"
	"github.com/edsrzf/mmap-go"
)
// type DL_CFG struct {
// 	per_thread  int
// 	// ststic path output
// }

type Chunk struct {
	index		int
	start		int64
	end			int64
	downloaded	int64
	status		string
	err			error
}

type DL_DATA struct {
	url			string
	name		string
	file		*os.File
	per_thread  int
	mmap		mmap.MMap
	size	    int64
	offset		int64
}

var unit = [6]string{"B", "KB", "MB", "GB", "TB", "PB"}

func (dl *DL_DATA) threaded_dl() error {
    // 1. Create and prepare file
    f, err := os.Create(dl.name)
    if err != nil {
        return fmt.Errorf("can't create file: %v", err)
    }
    dl.file = f
    defer dl.file.Close()

    // 2. Pre-allocate space
    err = dl.file.Truncate(dl.size)
    if err != nil {
        return fmt.Errorf("can't pre-allocate: %v", err)
    }

    // 3. Memory map the entire file
    mmap, err := mmap.Map(dl.file, mmap.RDWR, 0)
    if err != nil {
        return fmt.Errorf("can't mmap: %v", err)
    }
    dl.mmap = mmap
    defer dl.mmap.Unmap()

    // 4. Calculate chunk size and create chunks
    chunkSize := int64(10 * 1024 * 1024) // 10MB per chunk
    numChunks := (dl.size + chunkSize - 1) / chunkSize
    
    chunks := make([]Chunk, numChunks)
    for i := int64(0); i < numChunks; i++ {
        start := i * chunkSize
        end := start + chunkSize - 1
        if end >= dl.size {
            end = dl.size - 1
        }
        chunks[i] = Chunk{
            index:      int(i),
            start:      start,
            end:        end,
            downloaded: 0,
            status:     "pending",
        }
    }

    // 5. Download chunks in parallel
    var wg sync.WaitGroup
    sem := make(chan bool, dl.per_thread) // Limit concurrent downloads
    
    // Progress tracking
    var mu sync.Mutex
    var totalDownloaded int64
    
    // Start progress updater
    done := make(chan bool)
    go func() {
        ticker := time.NewTicker(500 * time.Millisecond)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                mu.Lock()
                downloaded := totalDownloaded
                mu.Unlock()
                dl.showProgress(downloaded)
            case <-done:
                return
            }
        }
    }()

    // Launch download for each chunk
    for i := range chunks {
        wg.Add(1)
        go func(chunk *Chunk) {
            sem <- true // Acquire semaphore
            defer func() { <-sem }() // Release semaphore
            defer wg.Done()
            
            err := dl.downloadChunk(chunk, &mu, &totalDownloaded)
            if err != nil {
                chunk.status = "error"
                chunk.err = err
            } else {
                chunk.status = "done"
            }
            
            // // Update total downloaded
            // mu.Lock()
            // totalDownloaded += chunk.downloaded
            // mu.Unlock()
        }(&chunks[i])
    }

    // Wait for all chunks to complete
    wg.Wait()
    done <- true // Stop progress updater
    dl.showProgress(totalDownloaded)
    fmt.Println() // New line after progress

    // Check for errors
    for _, chunk := range chunks {
        if chunk.status == "error" {
            return fmt.Errorf("chunk %d failed: %v", chunk.index, chunk.err)
        }
    }

    // Flush mmap to disk
    return dl.mmap.Flush()
}

func (dl *DL_DATA) downloadChunk(chunk *Chunk, mu *sync.Mutex, total *int64) error {
    easy := curl.EasyInit()
    if easy == nil {
        return fmt.Errorf("can't init curl for chunk %d", chunk.index)
    }
    defer easy.Cleanup()

    rangeHeader := fmt.Sprintf("%d-%d", chunk.start, chunk.end)
    easy.Setopt(curl.OPT_URL, dl.url)
    easy.Setopt(curl.OPT_RANGE, rangeHeader)
    easy.Setopt(curl.OPT_FOLLOWLOCATION, true)

    easy.Setopt(curl.OPT_WRITEFUNCTION, func(data []byte, userdata interface{}) bool {
        offset := chunk.start + chunk.downloaded
        end := offset + int64(len(data))
        if end > chunk.end+1 {
            return false
        }
        copy(dl.mmap[offset:end], data)
        chunk.downloaded += int64(len(data))

        // ✅ Update shared counter HERE, not after Perform()
        mu.Lock()
        *total += int64(len(data))
        mu.Unlock()

        return true
    })

	err := easy.Perform()
	if err != nil {
		return fmt.Errorf("chunk %d failed: %v", chunk.index, err)
	}

	// Verify server honored the range
	code, _ := easy.Getinfo(curl.INFO_RESPONSE_CODE)
	if code != 206 {
		return fmt.Errorf("chunk %d: got HTTP %d, expected 206", chunk.index, code)
	}

	// Verify we got exactly what we asked for
	expected := chunk.end - chunk.start + 1
	if chunk.downloaded != expected {
		return fmt.Errorf("chunk %d: got %d bytes, expected %d", chunk.index, chunk.downloaded, expected)
	}

    return nil
}

func (dl *DL_DATA) showProgress(downloaded int64) {
    percent := (float64(downloaded) / float64(dl.size)) * 100
    
    b_now := float64(downloaded)
    b_total := float64(dl.size)
    
    ind_now := 0
    ind_total := 0
    
    for b_now >= 1024 && ind_now < 5 {
        b_now /= 1024
        ind_now++
    }
    
    for b_total >= 1024 && ind_total < 5 {
        b_total /= 1024
        ind_total++
    }
    
    fmt.Printf("\r[%.2f%%] %.2f %s / %.2f %s    ", 
        percent, b_now, unit[ind_now], b_total, unit[ind_total])
}


// func single_dl(dl *DL_DATA) error {
// 	// init a file
// 	f, err := os.Create(dl.name)
// 	if err != nil {
// 		return fmt.Errorf("can't create file %v", err)
// 	}
// 	dl.file = f
// 	defer dl.file.Close()
//
// 	// pre size empty blocks
// 	err = dl.file.Truncate(dl.size)
// 	if err != nil {
// 		return fmt.Errorf("can't pre allocate space: %v", err)
// 	}
//
//
// 	// link with mmap
// 	mmap, err := mmap.Map(dl.file, mmap.RDWR, 0)
// 	if err != nil {
// 		return fmt.Errorf("can't memory-map file: %v", err)
// 	}
//
// 	dl.mmap = mmap
// 	defer dl.mmap.Unmap()
//
//
// 	// init a curl
// 	easy := curl.EasyInit()
// 	if easy == nil {
// 		return fmt.Errorf("can't init curl")
// 	}
//
// 	defer easy.Cleanup()
//
// 	easy.Setopt(curl.OPT_URL, dl.url) // link
//
// 	// write on a file using mmap
// 	easy.Setopt(curl.OPT_WRITEFUNCTION, func(data []byte, userdata interface{}) bool {
// 		dlPtr := userdata.(*DL_DATA)
// 		copy(dlPtr.mmap[dlPtr.offset:dlPtr.offset+int64(len(data))], data)
// 		dlPtr.offset += int64(len(data))
// 		return true
// 	})
//
// 	easy.Setopt(curl.OPT_WRITEDATA, dl)
//
// 	// progress
// 	easy.Setopt(curl.OPT_NOPROGRESS, false)
// 	easy.Setopt(curl.OPT_PROGRESSFUNCTION, progress_fn)
//
// 	err_easy := easy.Perform()
// 	if err_easy != nil {
// 		return fmt.Errorf("falied download: %v", err_easy)
// 	}
// 	return dl.mmap.Flush()
// }

func main() {
	// 1. arg checker
	if len(os.Args) < 2 {
		fmt.Printf("Usage: tego [LINK]")
		return
	}

	// 2. init context
	dl := DL_DATA{}

	if !arg_handle(&os.Args, &dl) {
		return
	}

	// default settings
	if dl.per_thread == 0 {
		dl.per_thread = 4
	}
	dl.url = os.Args[1]

	// 3. extract name from url
	if dl.name == "" {
		dl.name = extract_name(dl.url)
		if dl.name == "ERR" {
			fmt.Errorf("can't extract name!")
			return
		}
	}

	// if load_cfg(&dl.cfg) == 0 {
	// 	return
	// }

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

    fmt.Printf("Downloading: %s (%.2f MB)\n", dl.name, float64(dl.size)/(1024*1024))
    fmt.Printf("Using %d threads\n", dl.per_thread)

    // Start threaded download
    err = dl.threaded_dl()
    if err != nil {
        fmt.Printf("\nDownload failed: %v\n", err)
        return
    }
    
    fmt.Printf("\n✓ Download complete: %s\n", dl.name)
}
