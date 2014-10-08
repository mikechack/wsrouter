package main

import (
	"code.google.com/p/go.net/context"
	"code.google.com/p/go.net/websocket"
	"encoding/base64"
	//"encoding/json"
	"fmt"
	"log"
	//	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"sync"
	//"time"
)

type T struct {
	Msg   string
	From  string
	To    string
	Reply bool
	MsgId int64
}

type response struct {
	msgId   int64
	channel chan int64
}

type responses struct {
	sync.Mutex
	waiting          map[int64]response
	messagesReceived int
}

func (r *responses) isWaiting(id int64) bool {
	//fmt.Printf("check waiting\n")
	r.Lock()
	defer r.Unlock()
	_, ok := r.waiting[id]
	//fmt.Printf("check waiting done\n")
	return ok

}

type myLogger func(string, ...interface{}) (int, error)

func LogConditional(b bool, logger myLogger, s string, v ...interface{}) {
	if b {
		logger(s, v)
	}
}

//var R = responses{lock: new(sync.Mutex), waiting: make(map[int64]response)}
var R responses

var origin = "http://173.39.210.210/"
var srvurl = "ws://173.39.210.210:8080/echo/"

var printLog bool = true

var loopCount int = 0

const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "3333"
	CONN_TYPE = "tcp"
)

func replyWaiter(id int64, c chan int64) {
	select {
	case <-c:
		R.Lock()
		R.messagesReceived++
		if R.messagesReceived == loopCount {
			fmt.Printf("Messages Received Done = %v\n", R.messagesReceived)
		}
		delete(R.waiting, id)
		R.Unlock()
	}

}

func reader(ws *websocket.Conn, id string, cancel context.CancelFunc, c chan *T) {

	//var err error
	for {

		if !ws.IsClientConn() {
			log.Fatal("No Connection")
			cancel()
		}

		var data *T = new(T)
		err := websocket.JSON.Receive(ws, data)
		if err != nil {
			log.Fatal(err)
		}
		c <- data
		//LogConditional(printLog, fmt.Printf, "\nReceived message : %s\n", data.Msg)

	}
}

func writer(ws *websocket.Conn, me string, c chan *T) {

	conn, err := net.Dial("tcp", "173.39.210.64:2003")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for {

		var err error
		var b []byte
		if !ws.IsClientConn() {
			break
		}

		pkt := <-c
		b, err = base64.StdEncoding.DecodeString(pkt.Msg)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("Packet received in writer - let's send to Collectd (graphite currently) - %v\n", len(b))
		_, err = conn.Write(b)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	}

}

func main() {

	var server string

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	ctx, cancel = context.WithCancel(context.Background())

	runtime.GOMAXPROCS(1)

	server = os.Args[1]
	id := os.Args[2]

	LogConditional(printLog, fmt.Printf, "Client Id %s\n", id)
	//fmt.Printf("Client Id %s\n", id)

	var headers http.Header = make(http.Header)
	headers.Add("X-Client-ID", id)

	var srvurl = "ws://173.39.210.210:8080/echo/"

	origin = fmt.Sprintf("http://%s/", server)
	srvurl = fmt.Sprintf("ws://%s:8080/echo/?id=%s", server, id)

	u, err := url.Parse(srvurl)
	o, err := url.Parse(origin)
	ws, err := websocket.DialConfig(&websocket.Config{Location: u, Header: headers, Origin: o, Version: 13})

	if err != nil {
		log.Fatal(err)
	}

	c := make(chan *T)
	go writer(ws, id, c)
	go reader(ws, id, cancel, c)

	select {
	case <-ctx.Done():
	}

}
