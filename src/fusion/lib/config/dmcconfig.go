package config

import (
	_ "encoding/base64"
	"encoding/json"
	_ "fusion/lib/oauth"
	_ "fusion/lib/status"
	"log"
	"net/http"
	_ "net/url"
)

type dmcConfig struct {
	HostName        string `json:"hostName,omitempty"`
	Ip              string `json:"ip,omitempty"`
	SerialNumber    string `json:"serialNumber,omitempty"`
	SoftwareVersion string `json:"softwareVersion,omitempty"`
	DeviceType      string `json:"deviceType,omitempty"`
	HostOs          string `json:"hostOS,omitempty"`
}

var dmcconfig = dmcConfig{
	HostName:        "adasdf",
	Ip:              "1.1.1.1",
	SerialNumber:    "1234",
	SoftwareVersion: "V1",
	DeviceType:      "DMC",
	HostOs:          "CentosV7"}

func getDmcConfig(w http.ResponseWriter) {
	buf := make([]byte, 1500)
	w.Header().Set("Content-Type", "application/json")
	buf, err := json.Marshal(dmcconfig)
	if err != nil {
		log.Printf("Json error %v\n", err)
	}
	w.Write(buf)
}

func setDmcConfig(w http.ResponseWriter, r *http.Request) {
	buf := make([]byte, 1500)
	n, _ := r.Body.Read(buf)
	json.Unmarshal(buf[0:n], &dmcconfig)
	w.Header().Set("Content-Type", "application/json")
	buf, _ = json.Marshal(dmcconfig)
	w.Write(buf)
}
