package main

import "fmt"
import "log"
import "net/http"
import "code.google.com/p/go.net/websocket"
import "code.google.com/p/go.net/context"
import "registrar"
import "runtime"

//import "errors"

func transmitter(id string, ws *websocket.Conn, c chan *registrar.T, ctx context.Context, cancel context.CancelFunc) {
	var err error
	var data *registrar.T
	defer ws.Close()
	//defer close(c)
	defer cancel()
	defer registrar.RemoveConnection(id)
Loop:
	for {
		select {
		case data = <-c:
			err = websocket.JSON.Send(ws, *data)
			if err != nil {
				if !ws.IsClientConn() {
					log.Printf("transmitter closed\n")
				} else {
					log.Printf("transmitter error %v\n", err)
				}
				break Loop
			}
		case <-ctx.Done():
			log.Printf("transmitter closed")
			break Loop
		}
	}
}

func receiver(ws *websocket.Conn) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	ctx, cancel = context.WithCancel(context.Background())

	c := make(chan *registrar.T)
	defer ws.Close()
	//defer close(c)
	defer cancel()
	id := ws.Request().Header.Get("X-Client-ID")
	go transmitter(id, ws, c, ctx, cancel)
	err := registrar.NewConnection(id, c)
	if err != nil {
		fmt.Printf("New Connection Failed - Duplicate CLient Id %s\n", id)
		ws.Close()
	} else {
		for {
			var data *registrar.T = new(registrar.T)
			err = websocket.JSON.Receive(ws, data)
			if err != nil {
				log.Printf("echo handler error %v\n", err)
				break
			}
			data.From = id
			if data.Msg == "bye" {
				break
			}
			err = registrar.RouteMessage(data)
			if err != nil {
				log.Printf("Routing Error %v - %s\n", err, data.To)
			}
		}
		registrar.RemoveConnection(id)
	}
}

func validateConnectRequest(c *websocket.Config, r *http.Request) error {
	fmt.Printf("\n\nValidate Connection Request Client-ID- %v \n\n", r.Header.Get("X-Client-ID"))
	err := registrar.ValidateConnectionRequest(r.Header.Get("X-Client-ID"))
	log.Printf("Validate Connection Returned")
	if err != nil {
		log.Printf("Validate Connection Request Error %v\n", err)
	}
	return err
}

func httpRelay(w http.ResponseWriter, r *http.Request) {
	log.Printf("Http Relay - Method %s", r.Method)
}

func httpListener(c chan bool) {
	http.Handle("/echo/", websocket.Server{Handshake: validateConnectRequest, Handler: websocket.Handler(receiver)})
	http.Handle("/relay/", http.HandlerFunc(httpRelay))
	http.HandleFunc("/kill", func(x http.ResponseWriter, y *http.Request) { c <- true })
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func main() {

	fmt.Printf("Number of logical CPUs %x\n", runtime.NumCPU())
	chan1 := make(chan bool)
	go httpListener(chan1)
	select {
	case <-chan1:
		fmt.Printf("\n\nDetected context done\n\n")
	}
	fmt.Printf("goodbye\n")
}
