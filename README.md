# ProcEye

ProcEye is a monitoring tool that monitors your ressources per process. It gives you the possibility to follow your resource Consumption for each process. 
- Version: 0.1
- Platform: Linux
NOTE: This version is just work and needs much work to be a perfect one.

## Dependencies

- MongoDB
 - [Download and Install](https://www.mongodb.org/downloads)
- Libpcap
 - Install libpcap and libpcap-dev
- Go deps
```
$ go get github.com/gorilla/mux
$ go get gopkg.in/mgo.v2
```

## How to use?

Note: The default network interface is wlan0. You can change it
in "netowrk.go" file.
 
- You need permission to capture the traffic
```
$ sudo go run *.go
```
- Open http://127.0.0.1:8080

