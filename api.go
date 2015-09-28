package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

func StartAPI() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/system", SystemIndex)
	router.HandleFunc("/all/now", AllNowIndex)
	router.HandleFunc("/{pid}/{name}/cpu", CPUIndex)
	router.HandleFunc("/{pid}/{name}/mem", MemIndex)
	router.HandleFunc("/{pid}/{name}/net", NetIndex)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
	log.Fatal(http.ListenAndServe(":8080", router))
}

func SystemIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	var data []System
	err := connDB.Find(bson.M{"type": "system"}).Sort("-ts").Limit(30).All(&data)
	if err == nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			panic(err)
		}
	}
}

func AllNowIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	var data []System
	ts := time.Now().Unix() - 2
	err := connDB.Find(bson.M{"ts": ts}).Sort("-ts").All(&data)
	if err == nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			panic(err)
		}
	}
}

func CPUIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["pid"])
	name := string(vars["name"])
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	var data []CPU
	err := connDB.Find(bson.M{"type": "cpu", "pid": pid, "name": name}).Sort("-ts").Limit(60).All(&data)
	if err == nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			panic(err)
		}
	}
}

func MemIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["pid"])
	name := string(vars["name"])
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	var data []Memory
	err := connDB.Find(bson.M{"type": "memory", "pid": pid, "name": name}).Sort("-ts").Limit(60).All(&data)
	if err == nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			panic(err)
		}
	}
}

func NetIndex(w http.ResponseWriter, r *http.Request) {
	period := int64(60)
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["pid"])
	name := string(vars["name"])
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	var data []Network
	err := connDB.Find(bson.M{"type": "network", "pid": pid, "name": name}).Sort("-ts").Limit(int(period)).All(&data)
	if err == nil {
		now := time.Now().Unix() - 1
		last := now
		if len(data) > 0 {
			last = data[0].Ts
		} else {
			return
		}
		var i int64
		if now-last > period {
			d := []Network{}
			for i = 0; i < period; i++ {
				d = append(d, Network{data[0].Type, data[0].Pid, data[0].Name, 0, 0, (now - i), []*Sock{}, 0, 0})
			}
			if err := json.NewEncoder(w).Encode(d); err != nil {
				panic(err)
			}
			return
		}
		if now > last {
			for i = 0; i < (now - last); i++ {
				d := []Network{}
				d = append(d, Network{data[0].Type, data[0].Pid, data[0].Name, 0, 0, (now - i), []*Sock{}, 0, 0})
				data = append(data[:i], append(d, data[i:]...)...)
			}
		}
		diff := now - last
		if diff < 0 {
			diff = 0
		}
		for i = diff; i < period; i++ {
			if data[i].Ts < (last - i + (now - last)) {
				k := last - data[i].Ts + i - 1
				if k > period {
					k = period
				}
				for j := i; j < k; j++ {
					d := []Network{}
					d = append(d, Network{data[0].Type, data[0].Pid, data[0].Name, 0, 0, (data[0].Ts - j), []*Sock{}, 0, 0})
					data = append(data[:j], append(d, data[j:]...)...)
				}
				i = k
				last = data[i].Ts
			} else {
				last = data[i].Ts
			}
		}
		if len(data) > int(period) {
			data = data[:period]
		}
		if err := json.NewEncoder(w).Encode(data); err != nil {
			panic(err)
		}
	}
}
