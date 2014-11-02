package status

import (
	//"bytes"
	//"encoding/base64"
	"encoding/json"
	//"io"
	"log"
	"net/http"
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
