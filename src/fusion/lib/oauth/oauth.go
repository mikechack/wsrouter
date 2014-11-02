package oauth

/*
import "fmt"
import "log"
import "net/http"
import "runtime"
import "io"
import "bytes"
import "encoding/json"
import "encoding/base64"
*/

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type TokenResponse struct {
	BearerToken string
}

type MachineAccountHolder interface {
	GetName() string
	GetPassword() string
}

type machineAccount struct {
	Name      string `json:"name,omitempty"`
	Password  string `json:"password,omitempty"`
	AdminUser bool   `json:"adminUser,omitempty"`
}

func (ma machineAccount) GetName() string {
	return ma.GetName()
}

func (ma machineAccount) GetPassword() string {
	return ma.GetName()
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error {
	return nil
}

//https://idbrokerbts.webex.com/idb/token/81ba207e-c4a4-4f64-b8b8-a135dc2a96e5/v1/actions/GetBearerToken/invoke

func GetBearerTokenForMachineAccount(ma MachineAccountHolder) string {
	log.Printf("GetTokenForMachineAccount - name = %s\n", ma.GetName())
	log.Printf("GetTokenForMachineAccount - password = %s\n", ma.GetPassword())

	var creds = machineAccount{Name: ma.GetName(), Password: ma.GetPassword(), AdminUser: true}

	//buf := make([]byte, 5000)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://idbrokerbts.webex.com/idb/token/81ba207e-c4a4-4f64-b8b8-a135dc2a96e5/v1/actions/GetBearerToken/invoke", nil)
	req.Header.Set("Content-Type", "application/json")
	buf, err := json.Marshal(creds)
	req.Body = nopCloser{bytes.NewBuffer(buf)}
	if err != nil {
		log.Printf("Json error %v\n", err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	var token = TokenResponse{}
	body, err := ioutil.ReadAll(res.Body)
	//n, _ := res.Body.Read(buf[0:])
	//if err = json.Unmarshal(buf[0:n], &token); err != nil {
	if err = json.Unmarshal(body, &token); err != nil {
		log.Printf("buffer  %s\n", string(buf))
		log.Fatal(err)
	}

	//log.Printf("Here is the token %v", m.Access_token)
	res.Body.Close()

	return token.BearerToken
}
