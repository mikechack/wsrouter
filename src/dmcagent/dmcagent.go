package main

import (
	"fusion/lib/config"
	"fusion/lib/status"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func httpListener(port string, c chan bool) {

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}

func httpTlsListener(c chan bool) {
	log.Printf("About to listen on 443. Go to https://127.0.0.1:443/")
	err := http.ListenAndServeTLS(":443", "certs/server.pem", "certs/server.key", nil)
	if err != nil {
		log.Fatal(err)
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
	http.Handle("/status/", http.HandlerFunc(status.Status))
	http.Handle("/logon", http.HandlerFunc(config.CloudLogon))
	http.Handle("/getMachine", http.HandlerFunc(config.GetMachineAccount))
	http.Handle("/connect", http.HandlerFunc(config.Connect))
	http.Handle("/dmcConfig", http.HandlerFunc(config.DmcConfig))
	http.Handle("/getBearerForMachine", http.HandlerFunc(config.GetBearerForMachine))
	http.Handle("/getRealToken", http.HandlerFunc(config.GetRealToken))

	//debug related items - intended to aid in standalone, test environment
	http.Handle("/config/", http.HandlerFunc(config.Config))
	http.Handle("/logmeon", http.HandlerFunc(config.LogMeOn))
	http.Handle("/token", http.HandlerFunc(config.Token))
	chan1 := make(chan bool)
	go httpListener(port, chan1)
	go httpTlsListener(chan1)
	select {
	case <-chan1:
		log.Printf("\n\nDetected context done\n\n")
	}
	log.Printf("goodbye\n")
}
