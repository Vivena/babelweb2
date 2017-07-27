package ws

import (
	"babelweb2/parser"
	"bufio"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	delete = iota
	update = iota
)

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
func MCUpdates(updates chan parser.BabelUpdate, g *Listenergroup,
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
func HandleMessage(mess []byte, conn *websocket.Conn, telnet *telnetWarper,
	quit chan struct{}, rMess chan string) {
	
}

//Handler manage the websockets
func Handler(l *Listenergroup) http.Handler {
	var m2c Message
	m2c.Action = "client"

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	
	fn := func(w http.ResponseWriter, r *http.Request) {
		log.Println("New connection to a websocket")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Could not create the socket.", err)
			return
		}
		log.Println("    Sending the database to the new client")

		updates := NewListener()
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

			l.Push(updates)
			Db[router].M.Unlock()
		}		
		l.Flush(updates)
	}
	return http.HandlerFunc(fn)
}
