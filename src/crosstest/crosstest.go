package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type agentStatus struct {
	Connected bool `json:"connected"`
}

var currentStatus = agentStatus{Connected: false}

func GetState() bool {
	return currentStatus.Connected
}

func Status(w http.ResponseWriter, r *http.Request) {
	buf := make([]byte, 1500)

	r.ParseForm()
	log.Printf("DMC-Agent status %s\n", r.Method)

	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		buf, _ = json.Marshal(currentStatus)
		w.Write(buf)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func httpListener(port string, c chan bool) {

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}

func main() {

	rand.Seed(time.Now().Unix())

	var port string
	if len(os.Args) < 2 {
		port = "80"
	} else {
		port = os.Args[1]
	}

	log.Printf("Port number to listen on = %s\n", port)
	http.Handle("/status", http.HandlerFunc(Status))
	http.Handle("/status/", http.HandlerFunc(Status))

	chan1 := make(chan bool)
	go httpListener(port, chan1)
	select {
	case <-chan1:
		log.Printf("\n\nDetected context done\n\n")
	}
	log.Printf("goodbye\n")
}
