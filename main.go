package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"gopkg.in/mgo.v2"
)

var connDB *mgo.Collection

func workerCPU() {
	for e := range ChanCPU {
		err := connDB.Insert(&e)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func workerMem() {
	for e := range ChanMem {
		err := connDB.Insert(&e)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func workerNet() {
	for e := range ChanNet {
		err := connDB.Insert(&e)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func workerSystem() {
	for e := range ChanSystem {
		err := connDB.Insert(&e)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	session.SetMode(mgo.Monotonic, true)
	connDB = session.DB("proceye").C("ressources")

	go StartFeeds()
	go workerSystem()
	go workerCPU()
	go workerMem()
	go workerNet()
	go StartAPI()

	fmt.Println("[*] Started")
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
