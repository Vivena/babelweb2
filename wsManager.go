package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const chanelSize int = 1024
const port string = ":8080"
const htmlPage string = "test.html"
const (
	delete = iota
	update = iota
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

/*-----------------------------------------------------------------*/

func test(updates chan interface{}) {
	for {
		time.Sleep(1000000000)
		log.Println("test")
		updates <- "salut"
	}

}

/*-----------------------------------------------------------------*/

//GetMess gets messages sent by the client and redirect them to the mess chanel
func GetMess(conn *websocket.Conn, mess chan []byte) {
	_, message, err := conn.ReadMessage()
	if err != nil {
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

/*-----------------------------------------------------------------*/

//RootHandler load the html file when someone connect to the root of the server
func RootHandler(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile(htmlPage)
	if err != nil {
		log.Println("Could not open file.", err)
	}
	fmt.Fprintf(w, "%s", content)
}

//WsHandler manage the websockets
func WsHandler(updates chan interface{}) http.Handler { //TODO interface et non routeinfo
	fn := func(w http.ResponseWriter, r *http.Request) {

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Could not create the socket.", err)
			return
		}
		//TODO parcourt la base de donné et envois tout au client

		mess := make(chan []byte, chanelSize)
		go GetMess(conn, mess)
		for {
			//we wait for a new message from the client or from our chanel
			select {
			case lastUp := <-updates: //we got a new update on the chanel
				err := conn.WriteJSON(lastUp)
				if err != nil {
					log.Println(err)
				}

			case clientMessage := <-mess: //we got a message from the client
				HandleMessage(clientMessage)

			}
		}
	}
	return http.HandlerFunc(fn)
}

/*-----------------------------------------------------------------*/

func wsManager(updates chan interface{}) {
	//creation du chanel pour communiquer avec le reste du serv

	go test(updates)
	ws := WsHandler(updates)
	http.HandleFunc("/", RootHandler)
	http.Handle("/ws", ws)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}

}
