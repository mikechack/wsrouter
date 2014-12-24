package main

import (
	"github.com/streadway/amqp"

	"crypto/tls"
	//"encoding/json"
	"fmt"
	//"io/ioutil"
	"log"
	"net/http"
	"os"
	//"runtime/debug"
	"time"
)

type RegisterRequest struct {
	HostName        string `json:"hostName,omitempty"`
	Ip              string `json:"ip,omitempty"`
	SerialNumber    string `json:"serialNumber,omitempty"`
	SoftwareVersion string `json:"softwareVersion,omitempty"`
	DeviceType      string `json:"deviceType,omitempty"`
	HostOs          string `json:"hostOS,omitempty"`
	Token           string `json:"token,omitempty"`
}

var cert tls.Certificate
var exchangeName = "mediafusion"
var connection *amqp.Connection
var channel *amqp.Channel

//var url = "amqps://hlequgxb:P8jKTB_rYRdwRlrA26-gKAKHmWdZm4we@hyena.rmq.cloudamqp.com:5671/hlequgxb"

var url = "amqps://guest:guest@sj21lab-rabbitmq-1.cisco.com:5671/"

func init() {
	//defer debug.PrintStack()
	var err error
	cert, err = tls.LoadX509KeyPair("certs/client0-ca.crt", "certs/client0.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	tmpurl := os.Getenv("CLOUDAMQP_URL")
	if tmpurl != "" {
		url = tmpurl
	}

	connection, err := amqp.DialTLS(url, &tls.Config{Certificates: []tls.Certificate{cert}, PreferServerCipherSuites: true, InsecureSkipVerify: true})
	//connection, err := amqp.Dial(url)
	channel, _ = connection.Channel()
	durable := true
	autoDelete, noWait := false, false
	internal := false
	err = channel.ExchangeDeclarePassive(exchangeName, "topic", durable, autoDelete, internal, noWait, nil)
	if err != nil {
		log.Printf("Exchange does not exist: %s", err)
		channel, _ = connection.Channel()
		err = channel.ExchangeDeclare(exchangeName, "topic", durable, autoDelete, internal, noWait, nil)
		if err != nil {
			log.Fatalf("Could not create Exchange: %s", err)
		}
	}
}

/*
func register(w http.ResponseWriter, r *http.Request) {
	var reg = RegisterInfo{}
	body, err := ioutil.ReadAll(r.Body)
	if err = json.Unmarshal(body, &reg); err != nil {
		log.Fatal(err)
	}
	log.Printf("Register - Name       - %s\n", reg.SerialNumber)
	log.Printf("Register - Ip         - %s\n", reg.Ip)
	log.Printf("Register - Token      - %s\n", reg.Token)

	r.Body.Close()

}

func Register(w http.ResponseWriter, r *http.Request) {
	//buf := make([]byte, 1500)

	r.ParseForm()
	log.Printf("MF management - register device %s\n", r.Method)

	switch {
	case r.Method == "POST":
		register(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
*/

func httpListener(port string) {

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}

func main() {
	var err error
	defer connection.Close()
	//http.Handle("/register", http.HandlerFunc(Register))
	//go httpListener("8080")

	go func(con *amqp.Connection) {

		defer channel.Close()
		durable, exclusive := true, false
		autoDelete, noWait := false, true
		internal := false
		err = channel.ExchangeDeclarePassive("mediafusion", "topic", durable, autoDelete, internal, noWait, nil)
		if err != nil {
			log.Printf("Exchange does not exist %v\n", err)
		}

		q, _ := channel.QueueDeclare("signaling", durable, autoDelete, exclusive, noWait, nil)
		channel.QueueBind(q.Name, "dmc1", exchangeName, false, nil)
		autoAck, exclusive, noLocal, noWait := false, false, false, false
		messages, _ := channel.Consume(q.Name, "", autoAck, exclusive, noLocal, noWait, nil)
		multiAck := false
		for msg := range messages {

			fmt.Println("Body:", string(msg.Body), "Timestamp:", msg.Timestamp, "ConsumerTag:", msg.ConsumerTag)
			msg.Ack(multiAck)
		}
	}(connection)

	go func(con *amqp.Connection) {
		timer := time.NewTicker(1 * time.Second)

		for t := range timer.C {
			msg := amqp.Publishing{
				DeliveryMode: 1,
				Timestamp:    t,
				ContentType:  "text/plain",
				Body:         []byte("Hello world"),
			}
			mandatory, immediate := false, false
			channel.Publish("mediafusion", "dmc1", mandatory, immediate, msg)
		}
	}(connection)

	select {}
}
