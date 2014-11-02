package main

import "fmt"
import "log"
import "net/http"
import "runtime"
import "io"
import "bytes"
import "encoding/json"
import "encoding/base64"

//import "strings"

/*
var options = {
   host: 'idbroker.webex.com',
   port: 443,
   path: '/idb/oauth2/v1/access_token',
   method: 'POST',
   headers: headerTokenRequest
};
*/

var clientId = "C3268da01375c2f2579c608dd1fd0b34316e8581cfd5eb1b75dcf057f055590dc"
var clientSecret = "afcbada66b8267da73884671bbb9814718d8f9c6f06896c0dc6fbf1a2a7206b7"

var resourceId = "R32088c5530d474b8a9fe313e16e70a3ac59bd2cc06b411d9a80424baad81b0d1"
var resourceSecret = "5782e3d82fb10f5f1ff34c993ea7eff85550c600ae795974cc14d307cc7b1659"

var redirectUri = "https://mm.fusion.cisco.com"
var scopes = "mm-fusion:device_provisioning mm-fusion:device_connect"

type oauthCredentials struct {
	clientId     string
	clientSecret string
}

func (cred oauthCredentials) getAuthorization() string {
	result := "Basic " + base64.StdEncoding.EncodeToString([]byte(cred.clientId+":"+cred.clientSecret))
	return result
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error {
	return nil
}

type authTokenResponse struct {
	Token_type               string
	Access_token             string
	Refresh_token            string
	Expires_in               int
	Refresh_token_expires_in int
}

type tokenInfoResponse struct {
	User_id    string
	Cis_uuid   string
	Scope      []string
	Realm      string
	Expires_in int
}

func getMachineAccount(token string) {
	buf := make([]byte, 1500)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://identitybts.webex.com/organization/81ba207e-c4a4-4f64-b8b8-a135dc2a96e5/v1/Machines/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Body = nopCloser{bytes.NewBufferString("{\"name\":\"DMC1234567\", \"password\":\"+u_D{sZNW771ShE)28o>4f#so|3;z5FL\",\"entitlements\":[\"mm-fusion\"]}")}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Machine Account response status %s", res.Status)

	n, _ := res.Body.Read(buf[0:])
	//if err = json.Unmarshal(buf[0:n], &m); err != nil {
	//	log.Fatal(err)
	//}

	log.Printf("We got a response for machine account  = %v\n", string(buf[0:n]))

	res.Body.Close()
}

func getTokenInfo(token string) {
	buf := make([]byte, 1500)
	var m tokenInfoResponse
	var creds = oauthCredentials{resourceId, resourceSecret}

	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://idbrokerbts.webex.com/idb/oauth2/v1/tokeninfo", nil)
	req.Header.Set("Authorization", creds.getAuthorization())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Body = nopCloser{bytes.NewBufferString("access_token=" + token)}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Token Info response status %s", res.Status)

	n, _ := res.Body.Read(buf[0:])
	if err = json.Unmarshal(buf[0:n], &m); err != nil {
		log.Fatal(err)
	}
	log.Printf("We got a response for token info - user id  = %s\n", m.User_id)
	log.Printf("We got a response for token info - cis uuid = %s\n", m.Cis_uuid)
	log.Printf("We got a response for token info - realm    = %s\n", m.Realm)
	log.Printf("We got a response for token info - scope    = %v\n", m.Scope)

	res.Body.Close()
}

func getOauthToken(code string) {

	buf := make([]byte, 1500)
	var m authTokenResponse

	var creds = oauthCredentials{clientId, clientSecret}

	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://idbrokerbts.webex.com/idb/oauth2/v1/access_token", nil)
	req.Header.Set("Authorization", creds.getAuthorization())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Body = nopCloser{bytes.NewBufferString("grant_type=authorization_code&redirect_uri=https%3A%2F%2Fmm.fusion.cisco.com/token&code=" + code)}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	n, _ := res.Body.Read(buf[0:])
	if err = json.Unmarshal(buf[0:n], &m); err != nil {
		log.Fatal(err)
	}

	log.Printf("Here is the token %v", m.Access_token)
	res.Body.Close()

	getTokenInfo(m.Access_token)
	//getMachineAccount(m.Access_token)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("Http Login - Method %s", r.Method)
	//http.Redirect(w, r, "https://idbroker.webex.com/idb/oauth2/v1/authorize?response_type=code&client_id=C7d339ab61565beedbe4e76645ac530fde0e70a5a3ea2dc45676db8565904dd76&redirect_uri=https%3A%2F%2Fstratos.cisco.com%2Ftoken&scope=webexsquare%3Aget_conversation%20Identity%3ASCIM&state=random_string", 302)
	http.Redirect(w, r, "https://idbrokerbts.webex.com/idb/oauth2/v1/authorize?response_type=code&client_id=C3268da01375c2f2579c608dd1fd0b34316e8581cfd5eb1b75dcf057f055590dc&redirect_uri=https%3A%2F%2Fmm.fusion.cisco.com/token&scope=mm-fusion%3Adevice_provision%20mm-fusion%3Adevice_connect%20Identity%3ASCIM%20Identity%3AConfig%20Identity%3AOrganization&state=random_string", 302)

}

func authCode(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	code := r.FormValue("code")
	if code != "" {
		log.Printf("Http AuthCode - Method %s", r.Method)
		log.Printf("Http AuthCode - Code %s", code)
		getOauthToken(code)
	}
}

func httpTlsListener(c chan bool) {
	log.Printf("About to listen on 443. Go to https://127.0.0.1:443/")
	err := http.ListenAndServeTLS(":443", "certs/server.pem", "certs/server.key", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func httpListener(c chan bool) {

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}

func main() {

	//var creds = oauthCredentials{clientId, clientSecret}
	//fmt.Printf("auth = %s", creds.getAuthorization())
	http.Handle("/login/", http.HandlerFunc(loginPage))
	http.Handle("/token", http.HandlerFunc(authCode))
	fmt.Printf("Number of logical CPUs %x\n", runtime.NumCPU())
	chan1 := make(chan bool)
	go httpListener(chan1)
	go httpTlsListener(chan1)
	select {
	case <-chan1:
		fmt.Printf("\n\nDetected context done\n\n")
	}
	fmt.Printf("goodbye\n")
}
