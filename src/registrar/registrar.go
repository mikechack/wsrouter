package registrar

import "fmt"
import "sync"
import "errors"
import "time"

//var in = make(chan bool)

type T struct {
	Msg   string
	From  string
	To    string
	Reply bool
	MsgId int64
}

type Client map[string]chan *T

type connections struct {
	lock           sync.RWMutex
	NumConnections int
	Clients        Client
}

func (c connections) doesRouteExist(to string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	_, ok := c.Clients[to]
	return ok
}

var conn connections = connections{NumConnections: 0, Clients: make(Client)}

func (c *connections) Connected() int {
	return c.NumConnections
}

func (c *connections) IncNumConnections() int {
	c.NumConnections++
	return c.NumConnections
}

func (c *connections) DecNumConnections() int {
	c.NumConnections--
	return c.NumConnections
}

func NewConnection(id string, c chan *T) error {
	var err error
	conn.lock.Lock()
	defer conn.lock.Unlock()

	if _, ok := conn.Clients[id]; ok {
		err = errors.New("Duplicate Client Id")
	} else {
		err = nil
		conn.IncNumConnections()
		conn.Clients[id] = c
		fmt.Printf("Registrar: Total Connections %x\n", conn.Connected())
	}

	return err
}

func ValidateConnectionRequest(id string) error {
	var err error
	conn.lock.Lock()
	defer conn.lock.Unlock()

	if _, ok := conn.Clients[id]; ok {
		err = errors.New("Duplicate Client Id")
	} else {
		err = nil
	}

	return err
}

func RemoveConnection(id string) error {
	var err error
	conn.lock.Lock()
	defer conn.lock.Unlock()
	if _, ok := conn.Clients[id]; ok {
		delete(conn.Clients, id)
		conn.DecNumConnections()
		err = nil
	} else {
		err = errors.New("Client Id Does Not Exist")
	}

	fmt.Printf("Registrar: Total Connections %x\n", conn.Connected())

	return err
}

func RouteMessage(m *T) error {
	var err error
	if conn.doesRouteExist(m.To) {
		select {
		case conn.Clients[m.To] <- m:
		case <-time.After(5 * time.Second):
			fmt.Printf("Transmitter appears to be hung - %v\n", m.MsgId)
		}
	} else {
		err = errors.New("Destination Not Routable")
		panic("Oh my")
	}

	return err
}

func init() {

}
