package config

import (
	//"bytes"
	//"encoding/base64"
	"encoding/json"
	//"io"
	"fusion/lib/oauth"
	"fusion/lib/status"
	"log"
	"net/http"
	"net/url"
)

var clientId = "C3268da01375c2f2579c608dd1fd0b34316e8581cfd5eb1b75dcf057f055590dc"
var clientSecret = "afcbada66b8267da73884671bbb9814718d8f9c6f06896c0dc6fbf1a2a7206b7"

var redirectUri = url.QueryEscape("https://mm.fusion.cisco.com")
var scopes = "mm-fusion:device_provisioning mm-fusion:device_connect"
var ciUri = url.QueryEscape("https://idbrokerbts.webex.com/idb/oauth2/v1/authorize")

type MachineAccount struct {
	Name     string `json:"name,omitempty"`
	Password string `json:"password,omitempty"`
}

func (m MachineAccount) GetName() string {
	return m.Name
}

func (m MachineAccount) GetPassword() string {
	return m.Password
}

func (m *MachineAccount) copy(source MachineAccount) MachineAccount {
	m.Name = source.Name
	m.Password = source.Password
	return *m
}

var cannedmachine = MachineAccount{Name: "DMC1234567", Password: "+u_D{sZNW771ShE)28o>4f#so|3;z5FL"}
var machine = MachineAccount{}

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

func getMachineAccount(token string) MachineAccount {

	return machine.copy(cannedmachine)

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

func connect(r *http.Request) {
	log.Printf("DMC-Agent Connected - currently connected =  %v\n", status.GetState())
	buf := make([]byte, 1500)
	n, _ := r.Body.Read(buf)
	var machine = MachineAccount{}
	json.Unmarshal(buf[0:n], &machine)
	oauth.GetBearerTokenForMachineAccount(machine)
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
			getMachineAccount(token)
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
