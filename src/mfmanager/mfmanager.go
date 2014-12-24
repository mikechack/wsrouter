package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"mfmanager/doregister"
	"net/http"
	"os"
	"time"
)

var cert tls.Certificate

//var url = "amqps://hlequgxb:P8jKTB_rYRdwRlrA26-gKAKHmWdZm4we@hyena.rmq.cloudamqp.com/hlequgxb"
var url = "amqps://guest:guest@sj21lab-rabbitmq-ha.cisco.com:5671"
var exchangeName = "mediafusion"
var connection *amqp.Connection
var channel *amqp.Channel

func init() {
	var err error
	cert, err = tls.LoadX509KeyPair("certs/client0.crt", "certs/client0.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	tmpurl := os.Getenv("CLOUDAMQP_URL")
	if tmpurl != "" {
		url = tmpurl
	}
	connection, err := amqp.DialTLS(url, &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true})
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

func register(w http.ResponseWriter, r *http.Request) {
	var reg = doregister.RegisterRequest{}
	body, err := ioutil.ReadAll(r.Body)
	if err = json.Unmarshal(body, &reg); err != nil {
		log.Fatal(err)
	}
	log.Printf("Register - Serial     - %s\n", reg.SerialNumber)
	log.Printf("Register - Ip         - %s\n", reg.Ip)
	log.Printf("Register - Token      - %s\n", reg.Token)

	r.Body.Close()

	buf := make([]byte, 1500)
	response := doregister.RegisterResponse{RabbitURL: url, RabbitExchange: exchangeName, ClientTopic: "dmc1", MfServiceTopic: "mfservice"}
	w.Header().Set("Content-Type", "application/json")
	buf, err = json.Marshal(response)
	if err != nil {
		log.Printf("Json error %v\n", err)
	}
	w.Write(buf)

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

func httpListener(port string) {

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}

func main() {
	var err error
	defer connection.Close()
	http.Handle("/register", http.HandlerFunc(Register))
	go httpListener("8080")

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
		timer := time.NewTicker(60 * time.Second)

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

	doregister.Doregister()

	select {}
}
