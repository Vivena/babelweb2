package ws

import (
	"babelweb2/parser"
	"encoding/json"
	"log"
	"net/http"

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

//Message messages to send to the client via websocket
type Message map[string]interface{}

func weight(x int) int {
	x = (x & 0x5555) + ((x >> 1) & 0x5555)
	x = (x & 0x3333) + ((x >> 2) & 0x3333)
	x = (x & 0x0f0f) + ((x >> 4) & 0x0f0f)
	return (x & 0x00ff) + ((x >> 8) & 0x00ff)
}

/*-----------------------------------------------------------------*/

func getEntries() {

}

/*-----------------------------------------------------------------*/

//MCUpdates multicast updates sent by the routine comminicating with the routers
func MCUpdates(updates chan parser.BabelUpdate, g *Listenergroupe) {
	for {
		update, quit := <-updates
		if quit == false {
			log.Println("closing all chanels")
			close(globalClose)
			return
		}
		//lock()
		//bd.up()
		t := update.ToS()
		g.Iter(func(l *Listener) {
			l.conduct <- t
		})
		//unlock()
	}
}

//GetMess gets messages sent by the client and redirect them to the mess chanel
func GetMess(conn *websocket.Conn, mess chan []byte) {
	_, message, err := conn.ReadMessage()
	if err != nil {
		close(mess)
		log.Println(err)
		return
	}
	mess <- message

}

//HandleMessage handle messages receved from the client
func HandleMessage(mess []byte) {
	//TODO gere les message du client
	return
}

//Handler manage the websockets
func Handler(l *Listenergroupe) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		log.Println("bip")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Could not create the socket.", err)
			return
		}
		//TODO parcourt la base de donné et envois tout au client

		log.Println("New connection to a websocket")
		updates := NewListener()
		l.Push(updates)
		defer l.Flush(updates)
		mess := make(chan []byte, ChanelSize)
		go GetMess(conn, mess)
		for {
			//we wait for a new message from the client or from our chanel
			select {
			case lastUp := <-updates.conduct: //we got a new update on the chanel
				log.Println("\n test")
				j, jerr := json.Marshal(lastUp)
				if err != nil {
					log.Println(jerr)
				}
				test := Message{}
				json.Unmarshal(j, &test)
				log.Println("test1:", lastUp)
				log.Println("test2:", string(j))
				log.Println("test3:", test)

				err := conn.WriteJSON(j)
				if err != nil {
					log.Println(err)
				}

			case _, q := <-updates.quit:
				if q == false {

					return
				}

			case clientMessage, q := <-mess: //we got a message from the client
				if q == false {
					return
				}

				log.Println(clientMessage)

				HandleMessage(clientMessage)

			}
		}
	}
	return http.HandlerFunc(fn)
}
