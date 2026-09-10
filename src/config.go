package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// config content
// todo: add static path option
const default_cfg_content = `#tego config

# number of parallel threads (4 = default, range 1..32)
per_thread = 4

#later...`


func getConfigPath() (string, string) {
	home := os.Getenv("HOME") // ~
	path_parent := fmt.Sprintf("%s/.config/tego", home) // ~/.config/tego
	path_file := fmt.Sprintf("%s/tego.conf", path_parent) // ~/.config/tego/tego.conf

	return path_parent, path_file
}

func new_cfg(path_parent string, path_file string) {
	os.MkdirAll(path_parent, 0755) // make a dir
	f, _ := os.Create(path_file) // create a file
	f.WriteString(default_cfg_content) // write on a file
	f.Close()
	fmt.Println("Create new config is success")

}

func load_cfg(dl_cfg *DL_CFG) int {
	path_parent, path_file := getConfigPath()

	_, err := os.Stat(path_file) // check from metadata file
	if os.IsNotExist(err) { // file/path not found
		new_cfg(path_parent, path_file)

	}

	f, _ := os.Open(path_file)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") { continue } // skip empty and begin #

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {

			// check for per_thread
			if strings.TrimSpace(parts[0]) == "per_thread" {
				val, err := strconv.Atoi(strings.TrimSpace(parts[1]))
				if err != nil {
					fmt.Printf("a probloem in config, line per_thread\n")
					dl_cfg.per_thread = 4
				} else {
					dl_cfg.per_thread = val
					if checkPerThread(dl_cfg.per_thread) == 1 {
						fmt.Fprintln(os.Stderr, "thread limit exceeded, only (1..32)")
						return 0
					}
				}

			}

			// later ...
		}
	}

	f.Close()
	fmt.Printf("per_thread: %d\n", dl_cfg.per_thread)
	return 1
}
