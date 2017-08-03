package ws

import (
	"github.com/Vivena/babelweb2/parser"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

type node struct {
	m    *sync.Mutex
	desc *parser.BabelDesc
}

var nodes map[parser.Id]node

func Init() {
	nodes = make(map[parser.Id]node)
}

func AddDesc(d *parser.BabelDesc) {
	nodes[d.Id()] = node{desc: d, m: new(sync.Mutex)}
}

func GetDesc(id parser.Id) *parser.BabelDesc {
	return nodes[id].desc
}

func LockDesc(id parser.Id) {
	nodes[id].m.Lock()
}

func UnlockDesc(id parser.Id) {
	nodes[id].m.Unlock()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//Handler manage the websockets
func Handler(l *Listenergroup) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		log.Println("New connection to a websocket")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Could not create the socket.", err)
			return
		}
		log.Println("    Sending the database to the new client")
		for router := range nodes {
			nodes[router].m.Lock()
			nodes[router].desc.Iter(
				func(bu parser.BabelUpdate) error {
					sbu := bu.ToSUpdate()
					err := conn.WriteJSON(sbu)
					if err != nil {
						log.Println(err)
					}
					return err
				})
			nodes[router].m.Unlock()
		}
		updates := NewListener()
		l.Push(updates)
		defer l.Flush(updates)
		for {
			err := conn.WriteJSON(<-updates.Channel)
			if err != nil {
				log.Println(err)
			}
		}
	}
	return http.HandlerFunc(fn)
}
