package main

import (
	"fusion/lib/config"
	"fusion/lib/status"
	"log"
	"net/http"
)

func httpListener(c chan bool) {

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}

func main() {
	http.Handle("/status/", http.HandlerFunc(status.Status))
	http.Handle("/config/", http.HandlerFunc(config.Config))
	http.Handle("/logon", http.HandlerFunc(config.CloudLogon))
	http.Handle("/getMachine", http.HandlerFunc(config.GetMachineAccount))
	http.Handle("/connect", http.HandlerFunc(config.Connect))
	//http.Handle("/token/", http.HandlerFunc(authCode))
	chan1 := make(chan bool)
	go httpListener(chan1)
	select {
	case <-chan1:
		log.Printf("\n\nDetected context done\n\n")
	}
	log.Printf("goodbye\n")
}
