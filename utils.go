package main

import (
	"strings"
)


const (
	NAME_PROG = "tego"
)

func extract_name(url string) string{
	ind_url := strings.LastIndex(url, "/")

	if ind_url != -1 && ind_url < len(url) {
		return url[ind_url+1:]
	}
	return "ERR"
}
