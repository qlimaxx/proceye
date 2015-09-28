package main

// #cgo LDFLAGS: -lpcap
// #include "sniffex.c"
import "C"

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var device = "wlan0"
var networks = []*Network{}

//export newPacketHandler
func newPacketHandler(src *C.char, dst *C.char, length C.int) {
	source := C.GoString(src)
	destination := C.GoString(dst)
	ts := time.Now().UnixNano()
	for _, p := range networks {
		for _, s := range p.Socks {
			if s.Session[0] == source && s.Session[1] == destination {
				if (ts - p.Ts) > 1000000000 {
					p.Up = int(float32(p.LengthUp) / float32(1000) / (float32(ts-p.Ts) / float32(1000000000)))
					p.Down = int(float32(p.LengthDown) / float32(1000) / (float32(ts-p.Ts) / float32(1000000000)))
					p.Ts = ts
					p.LengthUp = 0
					p.LengthDown = 0
					if p.Up > 0 && p.Down > 0 {
						p.Ts = p.Ts / 1000000000
						ChanNet <- *p
					}
				}
				p.LengthUp += int(length)
			} else if s.Session[0] == destination && s.Session[1] == source {
				if (ts - p.Ts) > 1000000000 {
					p.Down = int(float32(p.LengthDown) / float32(1000) / (float32(ts-p.Ts) / float32(1000000000)))
					p.Up = int(float32(p.LengthUp) / float32(1000) / (float32(ts-p.Ts) / float32(1000000000)))
					p.Ts = ts
					p.LengthDown = 0
					p.LengthUp = 0
					if p.Up > 0 && p.Down > 0 {
						p.Ts = p.Ts / 1000000000
						ChanNet <- *p
					}
				}
				p.LengthDown += int(length)
			}
		}
	}
}

func readNetFile(fname string) []*Sock {
	socks := []*Sock{}
	fd, err := os.Open(fname)
	if err == nil {
		bf := bufio.NewReader(fd)
		bf.ReadLine()
		for {
			line, isPrefix, err := bf.ReadLine()
			if err == io.EOF {
				break
			}
			if err == nil && !isPrefix {
				inode, _ := strconv.Atoi(strings.Fields(string(line))[9])
				if inode != 0 {
					session := strings.Fields(string(line))[1:3]
					socks = append(socks, &Sock{uint64(inode), session})
				}
			}
		}
		fd.Close()
	}
	return socks
}

func getProcesses() {
	for {
		start := time.Now()
		inodes := readNetFile("/proc/net/tcp")
		var stat syscall.Stat_t
		fds, _ := filepath.Glob("/proc/[0-9]*/fd/[0-9]*")
		for _, fd := range fds {
			if err := syscall.Stat(fd, &stat); err == nil {
				if stat.Dev == 7 {
					for _, v := range inodes {
						if stat.Ino == v.Inode {
							found := false
							for _, p := range networks {
								pid, _ := strconv.Atoi(strings.Split(fd, "/")[2])
								if p.Pid == pid {
									found = true
									foundIno := false
									for _, ss := range p.Socks {
										if ss.Inode == v.Inode {
											foundIno = true
											break
										}
									}
									if !foundIno {
										p.Socks = append(p.Socks, v)
									}
									break
								}
							}
							if !found {
								cmd, _ := os.Open(strings.Join(append(strings.Split(fd, "fd")[0:1], "comm"), ""))
								name, _, _ := bufio.NewReader(cmd).ReadLine()
								cmd.Close()
								pid, _ := strconv.Atoi(strings.Split(fd, "/")[2])
								networks = append(networks, &Network{"network", pid, string(name), 0, 0, 0, []*Sock{v}, 0, 0})
							}
						}
					}
				}
			}
		}
		time.Sleep((1000000000 - time.Since(start)) * time.Nanosecond)
	}
}

func getNetwork() {
	var wg sync.WaitGroup
	go getProcesses()
	time.Sleep(500 * time.Millisecond)
	go C.sniff(C.CString(device))
	wg.Add(2)
	wg.Wait()
}
