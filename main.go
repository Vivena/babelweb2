package main

import (
	"babelweb2/parser"
	"babelweb2/ws"
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
)

const node = "[::1]:33123"

func Connection(updates chan parser.BabelUpdate, node string) {
	conn, err := net.Dial("tcp6", node)
	if err != nil {
		log.Println("node ", err)
		return
	}
	defer conn.Close()
	fmt.Fprintf(conn, "monitor\n")
	r := bufio.NewReader(conn)
	s := bufio.NewScanner(r)
	for {
		parser.Bd.Listen(s, updates)
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	updates := make(chan parser.BabelUpdate, ws.ChanelSize)
	parser.Bd = parser.NewBabelDesc()
	go Connection(updates, node)
	bcastGrp := ws.NewListenerGroupe()
	go ws.MCUpdates(updates, bcastGrp)
	ws := ws.Handler(bcastGrp)
	log.Println("manager lauched")
	http.Handle("/", http.FileServer(http.Dir("static/")))
	http.Handle("/style.css", http.FileServer(http.Dir("static/css/")))
	http.Handle("/initialize.js", http.FileServer(http.Dir("static/js")))
	http.Handle("/d3/d3.js", http.FileServer(http.Dir("static/js")))
	http.Handle("/ws", ws)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
	wg.Wait()
}
