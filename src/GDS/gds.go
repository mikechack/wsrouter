package main

import "fmt"
import "log"
import "net/http"
import "runtime"
import "math/rand"
import "time"
import "encoding/json"

//import "io"
//import "bytes"
//import "encoding/json"

type registration struct {
	user           string
	registrationId string
	ready          bool
	deviceId       string
	service        string
}

type redirectRequest struct {
	DeviceId string `json:"deviceId"`
	Service  string `json:"service"`
}

type redirectResponse struct {
	Status   int    `json:"status"`
	DeviceId string `json:"deviceId"`
	Service  string `json:"service"`
}

type pollResponse struct {
	Status   int    `json:"status"`
	DeviceId string `json:"deviceId"`
	Service  string `json:"service"`
}

type registerResponse struct {
	Status int    `json:"status"`
	User   string `json:"user"`
	Regid  string `json:"regid"`
}

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

type registrations map[string]registration

var reg = make(registrations)

func (r registrations) poll(user string, registrationId string) ([]byte, int) {
	registration, ok := r[user]
	if ok {
		if registration.registrationId == registrationId {
			log.Printf("Poll - Found a registration - ready = %v", registration.ready)
			if registration.ready {
				var response pollResponse
				response.DeviceId = registration.deviceId
				response.Service = registration.service
				response.Status = 200
				b, _ := json.Marshal(response)
				return b, 200
			}
		}
	}
	return []byte("{\"status\" : 200}"), 200
}

func (r registrations) redirect(user string, request redirectRequest) ([]byte, int) {
	registration, ok := r[user]
	if ok {
		log.Printf("Redirect - Found a registration - deviceId %s", request.DeviceId)
		registration.ready = true
		registration.deviceId = request.DeviceId
		registration.service = request.Service
		r[user] = registration
		var response redirectResponse
		response.DeviceId = request.DeviceId
		response.Service = request.Service
		response.Status = 200
		b, _ := json.Marshal(response)
		return b, 200
	} else {
		return []byte("{\"status\" : 404}"), 404
	}
}

func (r registrations) register(user string, regid string) ([]byte, int) {
	_, ok := r[user]
	if ok {
		log.Printf("Registration already exists")
		return []byte("{\"status : \"500}"), 500
	}

	log.Printf("Create Registration %s  %s", user, regid)
	r[user] = registration{user: user, registrationId: regid, ready: false}
	var response registerResponse
	response.Regid = regid
	response.User = user
	response.Status = 200
	b, _ := json.Marshal(response)
	return b, 200
}

func httpListener(c chan bool) {

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}

func genUser() string {

	letters := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

	b := make([]byte, 4)

	for i := 0; i < len(b); i++ {
		b[i] = letters[random(0, len(letters)-1)]
	}

	s := string(b[:])
	log.Printf("Generate user %s", s)
	return s
}

func register(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	log.Printf("GDS Registration %s", r.Method)
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "POST" {
		log.Printf("GDS Register - POST")
		u := genUser()
		regid := r.FormValue("regid")
		b, code := reg.register(u, regid)
		w.WriteHeader(code)
		w.Write(b)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func poll(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	log.Printf("GDS Registration poll %s", r.Method)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		log.Printf("GDS poll - GET")
		user := r.FormValue("user")
		regid := r.FormValue("regid")
		b, code := reg.poll(user, regid)
		w.WriteHeader(code)
		w.Write(b)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func redirect(w http.ResponseWriter, r *http.Request) {
	log.Printf("GDS Redirect %s", r.Method)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "PUT" {
		log.Printf("GDS Redirect - PUT")

		var req redirectRequest
		buf := make([]byte, 512)

		r.ParseForm()
		user := r.FormValue("user")
		n, err := r.Body.Read(buf)

		if err = json.Unmarshal(buf[0:n], &req); err != nil {
			log.Fatal(err)
		}

		b, code := reg.redirect(user, req)
		w.WriteHeader(code)
		w.Write(b)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {

	rand.Seed(time.Now().Unix())

	http.Handle("/register", http.HandlerFunc(register))
	http.Handle("/poll", http.HandlerFunc(poll))
	http.Handle("/redirect", http.HandlerFunc(redirect))
	fmt.Printf("Number of logical CPUs %x\n", runtime.NumCPU())
	chan1 := make(chan bool)
	go httpListener(chan1)
	//go httpTlsListener(chan1)
	select {
	case <-chan1:
		fmt.Printf("\n\nDetected context done\n\n")
	}
	fmt.Printf("goodbye\n")
}
