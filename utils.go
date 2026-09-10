package main

import (
	"strings"
)


const (
	NAME_PROG = "tego"
    HIDE_CURSOR = "\033[?25l"
    SHOW_CURSOR = "\033[?25h"
    // CLEAR_LINE  = "\r\033[K"
)

func extract_name(url string) string{
	ind_url := strings.LastIndex(url, "/")

	if ind_url != -1 && ind_url < len(url) {
		return url[ind_url+1:]
	}
	return "ERR"
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
