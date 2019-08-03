package ws

import (
	"log"
	"net/http"
	"sync"

	"github.com/Vivena/babelweb2/parser"
	"github.com/gorilla/websocket"
)

type nodelist struct {
	sync.Mutex
	nodes map[string]*parser.BabelDesc
}

var nodes nodelist

func Init() {
	nodes.nodes = make(map[string]*parser.BabelDesc)
}

func AddDesc(d *parser.BabelDesc) {
	nodes.Lock()
	nodes.nodes[d.ID] = d
	nodes.Unlock()
}

func RemoveDesc(id string) {
	nodes.Lock()
	delete(nodes.nodes, id)
	nodes.Unlock()
}

func GetDesc(id string) *parser.BabelDesc {
	nodes.Lock()
	defer nodes.Unlock()
	return nodes.nodes[id]
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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

		nodes.Lock()
		for _, node := range nodes.nodes {
			node.Iter(
				func(bu parser.BabelUpdate) error {
					sbu := bu.ToSUpdate()
					err := conn.WriteJSON(sbu)
					if err != nil {
						log.Println(err)
					}
					return err
				})
		}
		nodes.Unlock()
		updates := NewListener()
		l.Push(updates)

		// Ignore any data received on the websocket and detect
		// any errors.
		go func() {
			for {
				_, _, err := conn.NextReader()
				if err != nil {
					l.Flush(updates)
					conn.Close()
					break
				}
			}
		}()

		for {
			err := conn.WriteJSON(<-updates.Channel)
			if err != nil {
				log.Println(err)
			}
		}
	}
	return http.HandlerFunc(fn)
}
