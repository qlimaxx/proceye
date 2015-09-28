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

func getCPU() {
	go func() {
		for {
			start := time.Now()
			fd, err := os.Open("/proc/uptime")
			if err == nil {
				bf := bufio.NewReader(fd)
				buf, isPrefix, err := bf.ReadLine()
				if err == nil && !isPrefix {
					floatUptime, _ := strconv.ParseFloat(strings.Split(string(buf), " ")[0], 64)
					uptime := uint64(floatUptime)
					ps, _ := filepath.Glob("/proc/[0-9]*/stat")
					ts := time.Now().Unix()
					for _, p := range ps {
						fd, err := os.Open(p)
						if err == nil {
							bf := bufio.NewReader(fd)
							buf, isPrefix, err := bf.ReadLine()
							fd.Close()
							if err == nil && !isPrefix {
								stat := strings.Fields(string(buf))
								utime, _ := strconv.ParseUint(stat[13], 10, 64)
								stime, _ := strconv.ParseUint(stat[14], 10, 64)
								totalTime := utime + stime
								startTime, _ := strconv.ParseUint(stat[21], 10, 64)
								seconds := uptime - (startTime / 100)
								var usage float32
								if totalTime > 0 && seconds > 0 {
									usage = float32(float64(totalTime)/float64(seconds)/float64(scClkTck)) * 100
									if usage > 0.1 {
										usage = float32(int(usage*100)) / 100
										name := strings.Split(strings.Split(stat[1], "(")[1], ")")[0]
										pid, _ := strconv.Atoi(stat[0])
										ChanCPU <- CPU{"cpu", pid, name, usage, ts}
									}
								}
							}
						}
					}
				}
				fd.Close()
			}
			time.Sleep((1000000000 - time.Since(start)) * time.Nanosecond)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
