package ws

import (
	"log"
	"net/http"
	"sync"

	"github.com/Vivena/babelweb2/state"
	"github.com/gorilla/websocket"
)

type NodeList struct {
	sync.Mutex
	nodes    map[*state.BabelState]struct{}
	upgrader websocket.Upgrader
}

func NewNodeList() *NodeList {
	return &NodeList{
		nodes: make(map[*state.BabelState]struct{}),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (nl *NodeList) Add(s *state.BabelState) {
	nl.Lock()
	defer nl.Unlock()

	nl.nodes[s] = struct{}{}
}

func (nl *NodeList) Remove(s *state.BabelState) {
	nl.Lock()
	defer nl.Unlock()

	delete(nl.nodes, s)
}

func (nl *NodeList) Handler(l *Listenergroup) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		log.Println("New connection to a websocket")
		conn, err := nl.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Could not create the socket.", err)
			return
		}

		nl.Lock()
		for babel, _ := range nl.nodes {
			err := babel.Iter(func(t state.Transition) error {
				return conn.WriteJSON(t)
			})
			if err != nil {
				log.Println(err)
			}
		}
		nl.Unlock()

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
