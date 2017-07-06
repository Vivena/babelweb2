package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const port string = ":8080"
const htmlPage = "static/index.html"
const jsPage = "static/js/initialize.js"
const cssPage = "static/css/style.css"
const d3 = "static/js/d3/d3.js"

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

//MCUpdates multicast updates sent by the routine comminicating with the routers
func MCUpdates(updates chan interface{}, g *Listenergroupe) { //TODO changer string par le bon type
	for {
		update, quit := <-updates
		if quit == false {
			log.Println("closing all chanels")
			close(globalClose)
			return
		}
		g.Iter(func(l *Listener) {
			l.conduct <- update
		})

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

/*-----------------------------------------------------------------*/

//RootHandler load the html file when someone connect to the root of the server
func RootHandler(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile(htmlPage)
	if err != nil {
		log.Println("Could not open file.", err)
	}
	fmt.Fprintf(w, "%s", content)
}

//RootHandcss load the CSS file when someone connect to the root of the server
func RootHandcss(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile(cssPage)
	if err != nil {
		log.Println("Could not open file.", err)
	}
	fmt.Fprintf(w, "%s", content)
}

//RootHandinitialize load js file when someone connect to the root of the server
func RootHandinitialize(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile(jsPage)
	if err != nil {
		log.Println("Could not open file.", err)
	}
	fmt.Fprintf(w, "%s", content)
}
func d3Handler(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile(d3)
	if err != nil {
		log.Println("Could not open file.", err)
	}
	fmt.Fprintf(w, "%s", content)
}

//WsHandler manage the websockets
func WsHandler(l *Listenergroupe) http.Handler { //TODO interface et non routeinfo
	fn := func(w http.ResponseWriter, r *http.Request) {

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Could not create the socket.", err)
			return
		}
		//TODO parcourt la base de donné et envois tout au client

		log.Println("New connection to a websocket")
		updates := NewListener()
		l.Push(updates)
		defer l.Quit(updates)
		mess := make(chan []byte, chanelSize)
		go GetMess(conn, mess)
		for {
			//we wait for a new message from the client or from our chanel
			select {
			case lastUp := <-updates.conduct: //we got a new update on the chanel
				log.Println(lastUp)
				err := conn.WriteJSON(lastUp)
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

/*-----------------------------------------------------------------*/

func wsManager(updates chan interface{}) {
	//creation du chanel pour communiquer avec le reste du serv
	log.Println("test")
	go test(updates)
	bcastGrp := NewListenerGroupe()
	go MCUpdates(updates, bcastGrp)
	ws := WsHandler(bcastGrp)
	http.HandleFunc("/", RootHandler)
	http.HandleFunc("/style.css", RootHandcss)
	http.HandleFunc("/initialize.js", RootHandinitialize)
	http.HandleFunc("/d3/d3.js", d3Handler)

	http.Handle("/ws", ws)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}

}
