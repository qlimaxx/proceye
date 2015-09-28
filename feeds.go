package main

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

/*
#include <unistd.h>
#include <sys/types.h>
#include <pwd.h>
#include <stdlib.h>
*/
import "C"

var scClkTck C.long
var PageSize C.long
var TotalMemory uint64

var (
	ChanSystem chan System
	ChanCPU    chan CPU
	ChanMem    chan Memory
	ChanNet    chan Network
)

type System struct {
	Type string  `json:"type" bson:"type"`
	Pid  int     `json:"pid" bson:"pid"`
	Name string  `json:"name" bson:"name"`
	Cpu  float32 `json:"cpu" bson:"cpu"`
	Mem  float32 `json:"mem" bson:"mem"`
	Down int     `json:"down" bson:"down"`
	Up   int     `json:"up" bson:"up"`
	Ts   int64   `json:"ts" bson:"ts"`
}

type CPU struct {
	Type  string  `json:"type" bson:"type"`
	Pid   int     `json:"pid" bson:"pid"`
	Name  string  `json:"name" bson:"name"`
	Usage float32 `json:"cpu" bson:"cpu"`
	Ts    int64   `json:"ts" bson:"ts"`
}

type Memory struct {
	Type  string  `json:"type" bson:"type"`
	Pid   int     `json:"pid" bson:"pid"`
	Name  string  `json:"name" bson:"name"`
	Usage float32 `json:"mem" bson:"mem"`
	Ts    int64   `json:"ts" bson:"ts"`
}

type Network struct {
	Type       string  `json:"type" bson:"type"`
	Pid        int     `json:"pid" bson:"pid"`
	Name       string  `json:"name" bson:"name"`
	Up         int     `json:"up" bson:"up"`
	Down       int     `json:"down" bson:"down"`
	Ts         int64   `json:"ts" bson:"ts"`
	Socks      []*Sock `json:"-" bson:"-"`
	LengthUp   int     `json:"-" bson:"-"`
	LengthDown int     `json:"-" bson:"-"`
}

type Sock struct {
	Inode   uint64
	Session []string
}

func init() {
	scClkTck = C.sysconf(C._SC_CLK_TCK)
	PageSize = C.sysconf(C._SC_PAGESIZE)
	fd, err := os.Open("/proc/meminfo")
	if err == nil {
		bf := bufio.NewReader(fd)
		for {
			line, isPrefix, err := bf.ReadLine()
			if err == io.EOF {
				break
			}
			if err == nil && !isPrefix {
				if strings.HasPrefix(string(line), "MemTotal:") {
					fields := strings.Split(string(line), ":")
					if len(fields) != 2 {
						continue
					}
					value := strings.TrimSpace(fields[1])
					value = strings.Replace(value, " kB", "", -1)
					v, _ := strconv.ParseUint(value, 10, 64)
					TotalMemory = v
					break
				}
			}
		}
		fd.Close()
	}
}

func StartFeeds() {
	ChanSystem = make(chan System, 50)
	ChanCPU = make(chan CPU, 50)
	ChanMem = make(chan Memory, 50)
	ChanNet = make(chan Network, 50)

	go getSystem()
	go getNetwork()
	go getMemory()
	go getCPU()
}
