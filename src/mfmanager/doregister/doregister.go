package doregister

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	_ "regexp"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error {
	return nil
}

type RegisterRequest struct {
	HostName        string `json:"hostName,omitempty"`
	Ip              string `json:"ip,omitempty"`
	SerialNumber    string `json:"serialNumber,omitempty"`
	SoftwareVersion string `json:"softwareVersion,omitempty"`
	DeviceType      string `json:"deviceType,omitempty"`
	HostOs          string `json:"hostOS,omitempty"`
	Token           string `json:"token,omitempty"`
}

type RegisterResponse struct {
	RabbitURL      string `json:"rabbitURL,omitempty"`
	RabbitExchange string `json:"rabbitExchange,omitempty"`
	ClientTopic    string `json:"clientTopic,omitempty"`
	MfServiceTopic string `json:"MfServiceTopic,omitempty"`
}

func Doregister() {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "http://localhost:8080/register", nil)
	register := RegisterRequest{Token: "12345", SerialNumber: "11111111", Ip: "1.1.1.1"}

	buf, err := json.Marshal(register)

	req.Body = nopCloser{bytes.NewBuffer(buf)}
	if err != nil {
		log.Printf("Json error %v\n", err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	body = body

	res.Body.Close()

}
