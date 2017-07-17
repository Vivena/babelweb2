package main

import (
	"babelweb2/parser"
	"babelweb2/ws"
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

func flagsInit(node *string, bwPort *string) {
	var host, port, tempPort string
	flag.StringVar(&host, "h", "::1", "hostname (shorthand)")
	flag.StringVar(&port, "p", "33123", "port (shorthand)")
	flag.StringVar(&host, "host", "::1", "hostname")
	flag.StringVar(&port, "port", "33123", "port")
	flag.StringVar(&tempPort, "b", "8080", "babelweb Port (shorthanf)")
	flag.StringVar(&tempPort, "bwport", "8080", "babelweb Port ")
	flag.Parse()
	*node = "[" + host + "]:" + port
	*bwPort = ":" + tempPort
}

func Connection(updates chan parser.BabelUpdate, node string) {
	var conn net.Conn
	var err error
	for {
		log.Println("	Trying ", node)
		for {
			conn, err = net.Dial("tcp6", node)
			if err != nil {
				log.Println(err)
				time.Sleep(time.Second * 5)
			} else {
				break
			}
		}
		log.Println("	Connected to", node)
		fmt.Fprintf(conn, "monitor\n")
		r := bufio.NewReader(conn)
		s := parser.NewScanner(r)
		err = ws.Db.Bd.Listen(s, updates)
		conn.Close()
		log.Println("Connection closed")
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func main() {
	log.Println("	--------launching server--------")
	var node string
	var bwPort string
	flagsInit(&node, &bwPort)
	var wg sync.WaitGroup
	wg.Add(2)
	updates := make(chan parser.BabelUpdate, ws.ChanelSize)
	ws.Db.Bd = parser.NewBabelDesc()
	log.Println("connection to initial node:")
	go Connection(updates, node)
	bcastGrp := ws.NewListenerGroupe()
	go ws.MCUpdates(updates, bcastGrp)
	ws := ws.Handler(bcastGrp)
	http.Handle("/", http.FileServer(http.Dir("static/")))
	http.Handle("/style.css", http.FileServer(http.Dir("static/css/")))
	http.Handle("/initialize.js", http.FileServer(http.Dir("static/js")))
	http.Handle("/d3/d3.js", http.FileServer(http.Dir("static/js")))
	http.Handle("/ws", ws)
	err := http.ListenAndServe(bwPort, nil)
	if err != nil {
		log.Println(err)
		return
	}

	wg.Wait()
}
