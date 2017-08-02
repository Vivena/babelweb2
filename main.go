package main

import (
	"bufio"
	"container/list"
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
var Listconduct = list.New()
var static_root string
var ws_url string

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
		bd := parser.NewBabelDesc()
		bd.Fill(s)
		ws.AddDesc(bd)
		err := ws.Db[bd.Id()].Bd.Listen(s, updates)
		conn.Close()
		log.Println("Connection closed")
		if err != nil {
			log.Println(err)
			return
		}
		ws.Db[bd.Id()].M.Lock()
		err = ws.Db[bd.Id()].Bd.Clean(updates)
		ws.Db[bd.Id()].M.Unlock()
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func serveConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/javascript")
	fmt.Fprintf(w, "websocket_url = '%v'", ws_url)
}

func main() {
	var bwPort string

	flag.Var(&nodes, "node",
		"Babel node to connect to (may be repeated multiple times)")
	flag.StringVar(&bwPort, "http", ":8080", "web server address")
	flag.StringVar(&static_root, "static", "./static/",
		"directory with static files")
	flag.StringVar(&ws_url, "ws", "ws://localhost:8080",
		"location of the websocket")
	flag.Parse()

	ws.Init()
	log.Println("	--------launching server--------")

	updates := make(chan parser.BabelUpdate, ws.ChanelSize)
	defer close(updates)
	connection(updates, &bwPort)
	
	bcastGrp := ws.NewListenerGroup()
	handler := ws.Handler(bcastGrp)
	http.Handle("/", http.FileServer(http.Dir(static_root)))
	http.HandleFunc("/js/config.js", serveConfig)
	http.Handle("/ws", handler)
	err := http.ListenAndServe(bwPort, nil)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		update := <-updates
		if !(ws.Db[update.Id()].Bd.CheckUpdate(update)) {
			continue
		}
		ws.Db[update.Id()].M.Lock()
		err := ws.Db[update.Id()].Bd.Update(update)
		if err != nil {
			log.Println(err)
		}
		ws.Db[update.Id()].M.Unlock()
		t := update.ToSUpdate()
		bcastGrp.Iter(func(l *ws.Listener) {
			l.Conduct <- t
		})
	}
}
