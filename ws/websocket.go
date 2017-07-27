package ws

import (
	"babelweb2/parser"
	"bufio"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	delete = iota
	update = iota
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var Db map[parser.Id]dataBase

func Init() {
	Db = make(map[parser.Id]dataBase)
}

func AddDesc(bd *parser.BabelDesc) {
	Db[bd.Id()] = dataBase{Bd: bd, M: new(sync.Mutex)}
}

type Message struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

//Message messages to send to the client via websocket
// type Message map[string]interface{}

type dataBase struct {
	M *sync.Mutex
	Bd *parser.BabelDesc
}

type telnetWarper struct {
	sync.Mutex
	telnetcon net.Conn
}

//MCUpdates multicast updates sent by the routine comminicating with the routers
func MCUpdates(updates chan parser.BabelUpdate, g *Listenergroupe,
	wg *sync.WaitGroup) {
	wg.Add(1)
	for {
		update, quit := <-updates
		if !quit {
			log.Println("closing all channels")
			g.Iter(func(l *Listener) {
				close(l.conduct)
			})
			wg.Done()
			return
		}
		if !(Db[update.Id()].Bd.CheckUpdate(update)) {
			continue
		}
		// log.Println("sending : ", update)
		Db[update.Id()].M.Lock()
		err := Db[update.Id()].Bd.Update(update)
		if err != nil {
			log.Println(err)
		}
		Db[update.Id()].M.Unlock()
		t := update.ToSUpdate()
		g.Iter(func(l *Listener) {
			l.conduct <- t
		})
	}
}

//GetRouterMess gets messages sent by the current router and redirect them to
//the rMess channel
func GetRouterMess(telnet *telnetWarper, rMess chan string, quit chan struct{}) {
	s := bufio.NewScanner(bufio.NewReader(telnet.telnetcon))
	for {
		select {
		case <-quit:
			return
		default:
			s.Scan()
			if len(s.Text()) != 0 {
				log.Println(s.Text())
				rMess <- s.Text()
			}
		}
	}
}

//GetMess gets messages sent by the client and redirect them to the mess chanel
func GetMess(conn *websocket.Conn, mess chan []byte) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			close(mess)
			return
		}
		mess <- message
	}
}

//HandleMessage handle messages receved from the client
func HandleMessage(mess []byte, conn *websocket.Conn, telnet *telnetWarper, quit chan struct{}, rMess chan string) {
	var m2c Message
	m2c.Action = "client"
	var err error
	temp := strings.Split(string(mess), " ")
	log.Println(temp)

	if temp[0] == "connect" {
		//the ip we're asked to connect is valide
		log.Println("temp:", temp)
		if net.ParseIP(temp[1]) != nil {
			node := "[" + temp[1] + "]:33123"
			//it's the first connection
			if telnet.telnetcon != nil {
				log.Println("quit")
				quit <- struct{}{}
				log.Println("close?")
				telnet.telnetcon.Close()
				log.Println("close")

			}

			telnet.telnetcon, err = net.Dial("tcp6", node)

			if err != nil {
				log.Println("connection error")
				m2c.Message = "error"
				error := conn.WriteJSON(m2c)
				if error != nil {
					log.Println(err)
				}
			} else { //connection successfull
				go GetRouterMess(telnet, rMess, quit)
				m2c.Message = "connected " + node
				error := conn.WriteJSON(m2c)
				if error != nil {
					log.Println(err)
				}
			}
		} else {
			log.Println("not an IP")
			m2c.Message = "not an IP"
			error := conn.WriteJSON(m2c)
			if error != nil {
				log.Println(error)
			}
		}
	} else {
		if (telnet.telnetcon) == nil {
			log.Println("not connected")
		} else {
			log.Println("sending: ", string(mess))
			_, error := (telnet.telnetcon).Write(append(mess, byte('\n')))
			if error != nil {
				log.Println(error)
			}
		}
	}
	conn.WriteJSON(m2c)
	return
}

//Handler manage the websockets
func Handler(l *Listenergroupe) http.Handler {
	var m2c Message
	m2c.Action = "client"

	quit := make(chan struct{}, 2)
	fn := func(w http.ResponseWriter, r *http.Request) {

		log.Println("New connection to a websocket")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Could not create the socket.", err)
			return
		}
		log.Println("    Sending the database to the new client")
		for router := range Db {
			Db[router].M.Lock()
			Db[router].Bd.Iter(func(bu parser.BabelUpdate) error {
				sbu := bu.ToSUpdate()
				err := conn.WriteJSON(sbu)
				if err != nil {
					log.Println(err)
				}
				return err
			})
			Db[router].M.Unlock()
		}
		updates := NewListener()
		l.Push(updates)
		defer l.Flush(updates)

		var telnet telnetWarper
		messFromClient := make(chan []byte, ChanelSize)
		messFromRouter := make(chan string, 2)
		go GetMess(conn, messFromClient)

		for {
			//we wait for a new message from the client or from our channel
			select {
			case lastUp := <-updates.conduct: //we got a new update on the channel

				// log.Println("sending:\n", lastUp)

				err := conn.WriteJSON(lastUp)
				if err != nil {
					log.Println(err)
				}

				//we've got a message from the router
			case routerMessage := <-messFromRouter:
				log.Println("rmess: ", routerMessage)
				m2c.Message = routerMessage
				err := conn.WriteJSON(m2c)
				if err != nil {
					log.Println(err)
				}
				//we've got a message from the client
			case clientMessage, q := <-messFromClient:
				if q == false {
					return
				}

				HandleMessage(clientMessage, conn, &telnet, quit, messFromRouter)

				/*****************************************************************************/
			}
		}
	}
	return http.HandlerFunc(fn)
}
