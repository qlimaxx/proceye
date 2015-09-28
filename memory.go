package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

func getMemory() {
	go func() {
		for {
			start := time.Now()
			fnames, _ := filepath.Glob("/proc/[0-9]*/statm")
			ts := time.Now().Unix()
			for _, fname := range fnames {
				var rss, shared uint64
				fd, err := os.Open(fname)
				if err == nil {
					bf := bufio.NewReader(fd)
					line, isPrefix, err := bf.ReadLine()
					if err == nil && !isPrefix {
						data := strings.Fields(string(line))
						rss, _ = strconv.ParseUint(data[1], 10, 64)
						shared, _ = strconv.ParseUint(data[2], 10, 64)
					}
				}
				fd.Close()
				cmd, _ := os.Open(strings.Join(append(strings.Split(fname, "statm")[0:1], "comm"), ""))
				name, _, _ := bufio.NewReader(cmd).ReadLine()
				cmd.Close()
				pid, _ := strconv.Atoi(strings.Split(fname, "/")[2])
				usage := float32(int(float64((rss-shared)*uint64(PageSize/1024))/float64(TotalMemory)*10000)) / 100
				if usage > 0.01 {
					ChanMem <- Memory{"memory", pid, string(name), usage, ts}
				}
			}
			time.Sleep((1000000000 - time.Since(start)) * time.Nanosecond)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
