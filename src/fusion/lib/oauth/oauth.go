package oauth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var clientId = "C3268da01375c2f2579c608dd1fd0b34316e8581cfd5eb1b75dcf057f055590dc"
var clientSecret = "afcbada66b8267da73884671bbb9814718d8f9c6f06896c0dc6fbf1a2a7206b7"

type oauthCredentials struct {
	clientId     string
	clientSecret string
}

func (cred oauthCredentials) getAuthorization() string {
	result := "Basic " + base64.StdEncoding.EncodeToString([]byte(cred.clientId+":"+cred.clientSecret))
	return result
}

type BearerTokenResponse struct {
	BearerToken string `json:"BearerToken,omitempty"`
}

type TokenResponse struct {
	Access_token string `json:"access_token,omitempty"`
	Expires_in   int    `json:"expires_in,omitempty"`
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

func GetTokenForMachineAccount(bearer string) string {
	var creds = oauthCredentials{clientId, clientSecret}
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://idbrokerbts.webex.com/idb/oauth2/v1/access_token", nil)
	req.Header.Set("Authorization", creds.getAuthorization())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	v := url.Values{}
	v.Set("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer")
	v.Add("assertion", bearer)
	v.Add("scope", "mm-fusion:device_connect")
	req.Body = nopCloser{bytes.NewBufferString(v.Encode())}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	var token = TokenResponse{}
	body, err := ioutil.ReadAll(res.Body)
	if err = json.Unmarshal(body, &token); err != nil {
		log.Fatal(err)
	}

	log.Printf("Here is the token token token %s", token.Access_token)
	res.Body.Close()

	return ""
}

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

	var token = BearerTokenResponse{}
	body, err := ioutil.ReadAll(res.Body)
	if err = json.Unmarshal(body, &token); err != nil {
		log.Fatal(err)
	}

	//log.Printf("Here is the token %v", m.Access_token)
	res.Body.Close()

	return token.BearerToken
}
