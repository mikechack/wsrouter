package main

import (
	"code.google.com/p/go.net/context"
	"code.google.com/p/go.net/websocket"
	//"encoding/json"
	"fmt"
	"log"
	"math/rand"
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

func writer(ws *websocket.Conn, me string) {
	var i int
	var to string
	var msg string

	for {

		if !ws.IsClientConn() {
			break
		}

		time.Sleep(100 * time.Millisecond)
		fmt.Print("\n\nSend To: ")
		fmt.Scanf("%s", &to)
		fmt.Print("Message: ")
		fmt.Scanf("%s", &msg)
		for i, R.messagesReceived = 0, 0; i < loopCount; i++ {
			var data *T = new(T)
			//data.MsgId = time.Now().UnixNano()
			data.Reply = true
			data.MsgId = int64(i)
			data.To = to
			data.Msg = msg
			data.From = me
			r := response{msgId: data.MsgId, channel: make(chan int64)}
			R.Lock()
			R.waiting[data.MsgId] = r
			R.Unlock()
			go replyWaiter(data.MsgId, r.channel)
			err := websocket.JSON.Send(ws, data)
			if err != nil {
				log.Fatal(err)
			}
		}

	}

}

func main() {

	runtime.GOMAXPROCS(1)
	R = *new(responses)
	R.waiting = make(map[int64]response)
	rand.Seed(time.Now().UnixNano())
	var (
		ctx context.Context
	)

	ctx, _ = context.WithCancel(context.Background())

	id := os.Args[1]
	printLog = (os.Args[2] == "1")
	fmt.Sscanf(os.Args[3], "%d", &loopCount)

	LogConditional(printLog, fmt.Printf, "Client Id %s\n", id)
	//fmt.Printf("Client Id %s\n", id)

	var headers http.Header = make(http.Header)
	headers.Add("X-Client-ID", id)

	u, err := url.Parse(srvurl)
	o, err := url.Parse(origin)
	ws, err := websocket.DialConfig(&websocket.Config{Location: u, Header: headers, Origin: o, Version: 13})

	if err != nil {
		log.Fatal(err)
	}

	go writer(ws, id)
	go reader(ws, id)

	select {
	case <-ctx.Done():
	}

}
