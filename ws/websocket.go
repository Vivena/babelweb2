package ws

import (
	"github.com/Vivena/babelweb2/parser"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
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

type dataBase struct {
	M  *sync.Mutex
	Bd *parser.BabelDesc
}

//MCUpdates multicast updates sent by the routine comminicating with the routers
func MCUpdates(updates chan parser.BabelUpdate, g *Listenergroup) {
	for {
		update, quit := <-updates
		if !quit {
			log.Println("closing all channels")
			g.Iter(func(l *Listener) {
				close(l.conduct)
			})
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
		for {
			err := conn.WriteJSON(<-updates.conduct)
			if err != nil {
				log.Println(err)
			}
		}
	}
	return http.HandlerFunc(fn)
}
