package main

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

var cpuTotal uint64
var cpuIdle uint64

func calcCPU(s *System) {
	fd, err := os.Open("/proc/stat")
	if err == nil {
		bf := bufio.NewReader(fd)
		buf, isPrefix, err := bf.ReadLine()
		if err == nil && !isPrefix {
			array := strings.Fields(string(buf))[1:]
			var total uint64
			var idle uint64
			for i, e := range array {
				if i == 3 {
					v, _ := strconv.ParseUint(e, 10, 64)
					idle = v
					total += v
				} else {
					v, _ := strconv.ParseUint(e, 10, 64)
					total += v
				}
			}
			diffIdle := idle - cpuIdle
			diffTotal := total - cpuTotal
			s.Cpu = float32(int(float64(diffTotal-diffIdle)/float64(diffTotal)*10000.0)) / 100.0
			cpuIdle = idle
			cpuTotal = total
		}
		fd.Close()
	}
}

func calcMem(s *System) {
	fd, err := os.Open("/proc/meminfo")
	if err == nil {
		var total, available uint64
		bf := bufio.NewReader(fd)
		for {
			line, isPrefix, err := bf.ReadLine()
			if err == io.EOF {
				break
			}
			if err == nil && !isPrefix {
				fields := strings.Split(string(line), ":")
				if len(fields) != 2 {
					continue
				}
				key := strings.TrimSpace(fields[0])
				value := strings.TrimSpace(fields[1])
				value = strings.Replace(value, " kB", "", -1)
				v, _ := strconv.ParseUint(value, 10, 64)
				if key == "MemTotal" {
					total = v * 1024
				} else if key == "MemAvailable" {
					available = v * 1024
					break
				}
			}
		}
		fd.Close()
		s.Mem = float32(int(float64(total-available)/float64(total)*10000.0)) / 100.0
	}
}

var netDown uint64
var netUp uint64

func calcNet(s *System) {
	fd, err := os.Open("/proc/net/dev")
	if err == nil {
		bf := bufio.NewReader(fd)
		for {
			line, isPrefix, err := bf.ReadLine()
			if err == io.EOF {
				break
			}
			if err == nil && !isPrefix {
				if strings.Contains(string(line), device) {
					fields := strings.Fields(string(line))
					down, _ := strconv.ParseUint(fields[1], 10, 64)
					up, _ := strconv.ParseUint(fields[9], 10, 64)
					if netDown == 0 {
						netDown = down
						down = 0
					} else {
						tmp := down
						down = (down - netDown) / 1024
						netDown = tmp
					}
					if netUp == 0 {
						netUp = up
						up = 0
					} else {
						tmp := up
						up = (up - netUp) / 1024
						netUp = tmp
					}
					s.Down = int(down)
					s.Up = int(up)
					break
				}
			}
		}
		fd.Close()
	}
}

func getSystem() {
	system := System{}
	system.Type = "system"
	for {
		start := time.Now()
		system.Ts = time.Now().Unix()
		calcCPU(&system)
		calcMem(&system)
		calcNet(&system)
		ChanSystem <- system
		time.Sleep((1000000000 - time.Since(start)) * time.Nanosecond)
	}
}
