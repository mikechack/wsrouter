package oauth

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	_ "encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

import mrand "math/rand"

var clientId = "C71e2f13edd03a6307b9591f529345a90447d83814b6db35c26c18fc81044da2e"
var clientSecret = "be96c991c972f84696263449e4cdade3a12b4acec52996b3a529666f5a8b1237"

//var clientId = "C676b2fdd49118485d777ddd574baec1b6027b15fb0544401a3d3b903ad4076eb"
//var clientSecret = "773cdb97796bf23e4d87dbe1b07c6766044b02d949467c54a6e8f1e0454c2e4b"

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
	Access_token             string `json:"access_token,omitempty"`
	Expires_in               int    `json:"expires_in,omitempty"`
	Refresh_token            string `json:"refresh_token,omitempty"`
	Refresh_token_expires_in int    `json:"refresh_token_expires_in,omitempty"`
}

type MachineAccountHolder interface {
	GetName() string
	GetPassword() string
	GetOrganization() string
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

func random(min, max int) int {
	return mrand.Intn(max-min) + min
}

func GetSessionId(length int) string {

	alphabet := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

	b := make([]byte, length)

	for i := 0; i < len(b); i++ {
		b[i] = alphabet[random(0, len(alphabet)-1)]
	}

	s := string(b[:])
	log.Printf("Generate Sessionid %s", s)
	return s
}

func DecryptAesCBC(key []byte, crypted string) string {
	ciphertext, err := base64.URLEncoding.DecodeString(crypted)
	if err != nil {
		log.Printf("Decode String Error %s\n", err.Error())
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}

	iv := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	ciphertext = ciphertext[:len(ciphertext)-aes.BlockSize]

	// CBC mode always works in whole blocks.
	if len(ciphertext)%aes.BlockSize != 0 {
		panic("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	return string(ciphertext)

}

func GetTokenForMachineAccount(bearer string) string {
	//var creds = oauthCredentials{clientId, clientSecret}
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://idbroker.webex.com/idb/oauth2/v1/access_token", nil)
	//req.Header.Set("Authorization", creds.getAuthorization())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	v := url.Values{}
	v.Add("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer")
	v.Add("assertion", bearer)
	v.Add("client_id", clientId)
	v.Add("client_secret", clientSecret)
	v.Add("scope", "squared-fusion-mngmt:management squared-fusion-media:device_connect Identity:Config")
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

	log.Printf("Token for Machine Response Code      %s\n", res.Status)
	log.Printf("Token for Machine Here is the token  %s\n", token.Access_token)
	res.Body.Close()

	return token.Access_token
}

func GetBearerTokenForMachineAccount(ma MachineAccountHolder) string {
	log.Printf("GetTokenForMachineAccount - name = %s\n", ma.GetName())
	log.Printf("GetTokenForMachineAccount - password = %s\n", ma.GetPassword())

	var creds = machineAccount{Name: ma.GetName(), Password: ma.GetPassword(), AdminUser: true}

	//buf := make([]byte, 5000)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://idbroker.webex.com/idb/token/baab1ece-498c-452b-aea8-1a727413c818/v1/actions/GetBearerToken/invoke", nil)
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

func EncryptPKCS1v15(msg []byte) string {
	certPEMBlock, err := ioutil.ReadFile("./key.pem")
	if err != nil {
		log.Fatal("PEM error %v", err)
	}

	var keyDERBlock *pem.Block
	keyDERBlock, certPEMBlock = pem.Decode(certPEMBlock)

	var publickey *rsa.PublicKey
	if key, err := x509.ParsePKIXPublicKey(keyDERBlock.Bytes); err == nil {
		publickey = key.(*rsa.PublicKey)
		log.Printf("got a key")
	}

	encryptedmsg, err := rsa.EncryptPKCS1v15(rand.Reader, publickey, msg)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("OAEP encrypted [%s] to \n[%x]\n", string(msg), encryptedmsg)
	fmt.Println()

	return base64.URLEncoding.EncodeToString(encryptedmsg)
}
