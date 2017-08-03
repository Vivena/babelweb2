package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/Vivena/babelweb2/parser"
	"github.com/Vivena/babelweb2/ws"
	"log"
	"net"
	"net/http"
	"time"
)

type nodeslice []string

var nodes nodeslice
var staticRoot string
var wsURL string

func (i *nodeslice) String() string {
	return fmt.Sprintf("%s", *i)
}

func (i *nodeslice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func connection(updates chan parser.BabelUpdate, bwPort *string) {
	if len(nodes) == 0 {
		go connectionNode(updates, "[::1]:33123")
	} else {
		for i := 0; i < len(nodes); i++ {
			go connectionNode(updates, nodes[i])
		}
	}
}

func connectionNode(updates chan parser.BabelUpdate, node string) {
	var conn net.Conn
	var err error

	for {
		log.Println("	Trying ", node)
		exit := true
		for exit {
			conn, err = net.Dial("tcp6", node)
			if err != nil {
				log.Println(err)
				time.Sleep(time.Second * 5)
			} else {
				exit = false
			}
		}
		log.Println("	Connected to", node)
		fmt.Fprintf(conn, "monitor\n")
		r := bufio.NewReader(conn)
		s := parser.NewScanner(r)
		desc := parser.NewBabelDesc()
		desc.Fill(s)
		ws.AddDesc(desc)
		err := desc.Listen(s, updates)
		conn.Close()
		log.Println("Connection closed")
		if err != nil {
			log.Println(err)
			return
		}
		ws.LockDesc(desc.Id())
		err = desc.Clean(updates)
		ws.UnlockDesc(desc.Id())
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func serveConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/javascript")
	fmt.Fprintf(w, "websocket_url = '%v'", wsURL)
}

func main() {
	var bwPort string

	flag.Var(&nodes, "node",
		"Babel node to connect to (may be repeated multiple times)")
	flag.StringVar(&bwPort, "http", ":8080", "web server address")
	flag.StringVar(&staticRoot, "static", "./static/",
		"directory with static files")
	flag.StringVar(&wsURL, "ws", "ws://localhost:8080",
		"location of the websocket")
	flag.Parse()

	ws.Init()
	log.Println("	--------launching server--------")

	updates := make(chan parser.BabelUpdate, 1024)
	defer close(updates)
	connection(updates, &bwPort)

	bcastGrp := ws.NewListenerGroup()
	handler := ws.Handler(bcastGrp)
	http.Handle("/", http.FileServer(http.Dir(staticRoot)))
	http.HandleFunc("/js/config.js", serveConfig)
	http.Handle("/ws", handler)
	err := http.ListenAndServe(bwPort, nil)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		update := <-updates
		desc := ws.GetDesc(update.Id())
		if !(desc.CheckUpdate(update)) {
			continue
		}
		ws.LockDesc(update.Id())
		err := desc.Update(update)
		if err != nil {
			log.Println(err)
		}
		ws.LockDesc(update.Id())
		t := update.ToSUpdate()
		bcastGrp.Iter(func(l *ws.Listener) {
			l.Conduct <- t
		})
	}
}
