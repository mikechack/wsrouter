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
	"time"
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

func reader(ws *websocket.Conn, id string) {

	//var err error
	for {

		if !ws.IsClientConn() {
			log.Fatal("No Connection")
			break
		}

		var data *T = new(T)
		err := websocket.JSON.Receive(ws, data)
		if err != nil {
			log.Fatal(err)
		}

		LogConditional(printLog, fmt.Printf, "\nReceived message : %s\n", data.Msg)
		if data.Reply {
			if data.From != id {
				LogConditional(printLog, fmt.Printf, "Need to Reply to %v\n", data.From)
				data.To = data.From
				data.From = id
				data.Reply = false
				data.Msg = fmt.Sprintf("%s%s", "Back at you ------- ", data.Msg)
				err := websocket.JSON.Send(ws, *data)
				if err != nil {
					log.Fatal(err)
				}
			}
		} else {
			if R.isWaiting(data.MsgId) {
				R.Lock()
				channel := R.waiting[data.MsgId].channel
				R.Unlock()
				channel <- data.MsgId
			}
		}

	}
}

func writer(ws *websocket.Conn, me string, c chan []byte) {

	for {

		if !ws.IsClientConn() {
			break
		}

		b := <-c
		msg := base64.StdEncoding.EncodeToString(b)
		fmt.Printf("Packet received in writer - length %v\n", len(b))
		//fmt.Println(msg)

		var data *T = new(T)
		data.MsgId = time.Now().UnixNano()
		data.Reply = false
		data.To = "collectd"
		data.Msg = msg
		data.From = me
		err := websocket.JSON.Send(ws, data)
		if err != nil {
			log.Fatal(err)
		}

	}

}

const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "3333"
	CONN_TYPE = "tcp"
)

func collectSyslog(c chan []byte) {
	var buf [1024]byte
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:5555")
	if err != nil {
		fmt.Println("Error resolve address:", err.Error())
		os.Exit(1)
	}
	sock, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	for {
		rlen, _, _ := sock.ReadFromUDP(buf[:])
		fmt.Printf("Got Syslog Message\n%s\n", buf[0:rlen])
	}
}

func collectdTcp(cancel context.CancelFunc, c chan []byte) {
	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn, cancel, c)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn, cancel context.CancelFunc, c chan []byte) {
	// Make a buffer to hold incoming data.
	// Read the incoming connection into the buffer.
	for {
		buf := make([]byte, 1500)
		n, err := conn.Read(buf[0:])
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			//cancel()
			conn.Close()
			break
		}
		fmt.Printf("Packet received - buf address %v", &buf[0])
		c <- buf[0:n]
	}
}

func collectd(ctx context.Context) {

	laddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:3333")
	if err != nil {
		log.Fatalln("fatal: failed to resolve address", err)
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		log.Fatalln("fatal: failed to listen", err)
	}
	for {
		buf := make([]byte, 1452)
		n, err := conn.Read(buf[:])
		if err != nil {
			log.Println("error: Failed to recieve packet", err)
		} else {
			log.Println("Received packet - length", n)
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

	c := make(chan []byte)

	go collectdTcp(cancel, c)
	go collectSyslog(c)
	go writer(ws, id, c)
	//go reader(ws, id)

	select {
	case <-ctx.Done():
	}

}
