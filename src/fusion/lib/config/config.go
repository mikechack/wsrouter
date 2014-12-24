package config

import (
	"crypto/tls"
	_ "encoding/hex"
	"encoding/json"
	"fusion/lib/oauth"
	"fusion/lib/status"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

var cert tls.Certificate
var clientId = "C3268da01375c2f2579c608dd1fd0b34316e8581cfd5eb1b75dcf057f055590dc"
var clientSecret = "afcbada66b8267da73884671bbb9814718d8f9c6f06896c0dc6fbf1a2a7206b7"

var redirectUri = url.QueryEscape("https://mm.fusion.cisco.com")
var scopes = "mm-fusion:device_provisioning mm-fusion:device_connect"
var ciUri = "https://idbrokerbts.webex.com/idb/oauth2/v1/authorize?response_type=code&client_id=C3268da01375c2f2579c608dd1fd0b34316e8581cfd5eb1b75dcf057f055590dc&redirect_uri=https%3A%2F%2Fmm.fusion.cisco.com/token&scope=mm-fusion%3Adevice_provision%20mm-fusion%3Adevice_connect%20Identity%3ASCIM%20Identity%3AConfig%20Identity%3AOrganization&state=random_string"

type MachineAccount struct {
	Username        string `json:"username,omitempty"`
	Password        string `json:"password,omitempty"`
	Location        string `json:"location,omitempty"`
	Organization_id string `json:"organization_id,omitempty"`
	Account_id      string `json:"account_id,omitempty"`
}

func (m MachineAccount) GetName() string {
	return m.Username
}

func (m MachineAccount) GetPassword() string {
	return m.Password
}

func (m MachineAccount) GetOrganization() string {
	return m.Password
}

func (m *MachineAccount) copy(source MachineAccount) MachineAccount {
	m.Username = source.Username
	m.Password = source.Password
	return *m
}

var cannedmachine = MachineAccount{Username: "fusion-mgmnt-dcd85bf6-2f05-4b47-b85b-63f44f785513", Password: "aaBB12$99a50ccc-0ee0-4d3e-8395-522043c56e4a", Location: "https://identity.webex.com/organization/baab1ece-498c-452b-aea8-1a727413c818/v1/Machines/412c5795-d8cf-49cf-9e67-504fa2e045d8", Organization_id: "baab1ece-498c-452b-aea8-1a727413c818"}

var machine = MachineAccount{}
var sessionId string
var tempToken string
var bearerToken string
var token string

type agentConfig struct {
	ClientId     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
	RedirectUri  string `json:"redirectUri,omitempty"`
	Scopes       string `json:"scopes,omitempty"`
	Uri          string `json:"uri,omitempty"`
}

var config = agentConfig{
	ClientId:     clientId,
	ClientSecret: clientSecret,
	RedirectUri:  redirectUri,
	Scopes:       scopes,
	Uri:          ciUri}

func (c *agentConfig) constructUri() string {
	return config.Uri
}

func getMachineAccount_dummy(token string) MachineAccount {

	return machine.copy(cannedmachine)

}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error {
	return nil
}

func getConfig(w http.ResponseWriter) {
	buf := make([]byte, 1500)
	w.Header().Set("Content-Type", "application/json")
	buf, err := json.Marshal(config)
	if err != nil {
		log.Printf("Json error %v\n", err)
	}
	w.Write(buf)
}

func setConfig(w http.ResponseWriter, r *http.Request) {
	buf := make([]byte, 1500)
	n, _ := r.Body.Read(buf)
	json.Unmarshal(buf[0:n], &config)
	w.Header().Set("Content-Type", "application/json")
	buf, _ = json.Marshal(config)
	w.Write(buf)
}

func saveMachineAccount(fname string, machine MachineAccount) {
	buf, _ := json.Marshal(machine)
	ioutil.WriteFile(fname, buf, os.ModePerm)
}

func connect(r *http.Request) {
	log.Printf("DMC-Agent Connected - currently connected =  %v\n", status.GetState())
	buf := make([]byte, 1500)
	n, _ := r.Body.Read(buf)
	var machine = MachineAccount{}
	json.Unmarshal(buf[0:n], &machine)
	bearerToken := oauth.GetBearerTokenForMachineAccount(machine)
	oauth.GetTokenForMachineAccount(bearerToken)
}

func Connect(w http.ResponseWriter, r *http.Request) {
	log.Printf("DMC-Agent Connect %s\n", r.Method)

	switch {
	case r.Method == "POST":
		connect(r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func GetMachineAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		r.ParseForm()
		token := r.FormValue("token")
		if token != "" {
			buf := make([]byte, 1500)
			log.Printf("Get machine Account - token = %s\n", token)
			getMachineAccount_dummy(token)
			w.Header().Set("Content-Type", "application/json")
			buf, err := json.Marshal(machine)
			if err != nil {
				log.Printf("Json error %v\n", err)
			}
			w.Write(buf)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func CloudLogon(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		buf := make([]byte, 1500)
		r.ParseForm()
		w.Header().Set("Content-Type", "application/json")
		buf, err := json.Marshal(agentConfig{Uri: config.constructUri()})
		if err != nil {
			log.Printf("Json error %v\n", err)
		}
		w.Write(buf)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func Token(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		key := []byte("12345678901234567890123456789012")
		r.ParseForm()
		token := r.FormValue("access_token")
		sId := r.FormValue("session_id")
		log.Printf("Token crypted - token      - %s\n", token)
		log.Printf("Token crypted - session_id - %s\n", sId)
		sId = oauth.DecryptAesCBC(key, sId)
		log.Printf("Decrypted Session = %s\n", sId)
		tempToken = oauth.DecryptAesCBC(key, token)
		log.Printf("Decrypted Token   = %s\n", tempToken)
		ma := getMachineAccount(tempToken, sId)
		machine = ma
		bearerToken = oauth.GetBearerTokenForMachineAccount(machine)
		token = oauth.GetTokenForMachineAccount(bearerToken)
		fmsRegisterDevice(token)
		//http.Redirect(w, r, "https://int-admin.wbx2.com/#/login", 302)

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func Config(w http.ResponseWriter, r *http.Request) {
	//buf := make([]byte, 1500)

	r.ParseForm()
	log.Printf("DMC-Agent config %s\n", r.Method)

	switch {
	case r.Method == "GET":
		getConfig(w)
	case r.Method == "PUT":
		setConfig(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func DmcConfig(w http.ResponseWriter, r *http.Request) {
	log.Printf("DMC-Agent config %s\n", r.Method)

	switch {
	case r.Method == "GET":
		getDmcConfig(w)
	case r.Method == "PUT":
		setDmcConfig(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func LogMeOn(w http.ResponseWriter, r *http.Request) {
	log.Printf("Http LogMeOn - Method %s", r.Method)
	sessionId = oauth.GetSessionId(48)
	log.Printf("Session Id = %s", sessionId)
	secret := "12345678901234567890123456789012"
	uri := "{\"redirect_uri\":\"https://localhost/token\",\"session_id\":\"" + sessionId + "\",\"box_name\":\"\",\"shared_secret\":\"" + secret + "\"}"
	finalRedirect := oauth.EncryptPKCS1v15([]byte(uri))
	//http.Redirect(w, r, "https://idbroker.webex.com/idb/oauth2/v1/authorize?response_type=code&client_id=C7d339ab61565beedbe4e76645ac530fde0e70a5a3ea2dc45676db8565904dd76&redirect_uri=https%3A%2F%2Fstratos.cisco.com%2Ftoken&scope=webexsquare%3Aget_conversation%20Identity%3ASCIM&state=random_string", 302)
	//http.Redirect(w, r, "https://idbroker.webex.com/idb/oauth2/v1/authorize?response_type=token&client_id=C0cd283dc5b7d8bd5929a825324c74b4d2755d14cf52eb4b256f9a63bec15fce8&redirect_uri=https%3A%2F%2Fhercules.hitest.huron-dev.com%2Ffuse_redirect&scope=Identity%3ASCIM%20Identity%3AOrganization&state="+finalRedirect, 302)
	http.Redirect(w, r, "https://idbroker.webex.com/idb/oauth2/v1/authorize?response_type=token&client_id=C0cd283dc5b7d8bd5929a825324c74b4d2755d14cf52eb4b256f9a63bec15fce8&redirect_uri=https%3A%2F%2Fhercules.ladidadi.org%2Ffuse_redirect&scope=Identity%3ASCIM%20Identity%3AOrganization&state="+finalRedirect, 302)

}

func GetBearerForMachine(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		bearerToken = oauth.GetBearerTokenForMachineAccount(cannedmachine)
		log.Printf("Bearer Token = %s\n", bearerToken)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func GetRealToken(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		token = oauth.GetTokenForMachineAccount(bearerToken)
		log.Printf("Real token = %s\n", token)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func init() {
	var err error
	cert, err = tls.LoadX509KeyPair("certs/client0.crt", "certs/client0.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
}
