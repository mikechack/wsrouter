package config

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
)

type jsonFmsRegisterDevice struct {
	Serial         string `json:"serial,omitempty"`
	Host_name      string `json:"host_name,omitempty"`
	Connector_type string `json:"connector_type,omitempty"`
	Version        string `json:"serial,omitempty"`
}

func fmsRegisterDevice(token string) {
	regInfo := jsonFmsRegisterDevice{Serial: "12341234", Host_name: "dmc1.mfusion1webx.com", Connector_type: "dmc_management_connector", Version: "1.0"}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, _ := http.NewRequest("POST", "https://hercules.ladidadi.org/v1/connectors", nil)
	//req, _ := http.NewRequest("POST", "https://hercules.hitest.huron-dev.com/v1/machine_accounts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	buf, err := json.Marshal(regInfo)
	log.Printf("Hercules-GetMA - json - %s\n", buf)
	req.Body = nopCloser{bytes.NewBuffer(buf)}
	if err != nil {
		log.Printf("Json error %v\n", err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Register Device - Status - %s\n", res.Status)

	/*
		var ma = MachineAccount{}
		body, err := ioutil.ReadAll(res.Body)
		if err = json.Unmarshal(body, &ma); err != nil {
			log.Fatal(err)
		}
	*/

	res.Body.Close()
}
