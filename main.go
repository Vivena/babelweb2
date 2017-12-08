package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/Vivena/babelweb2/parser"
	"github.com/Vivena/babelweb2/ws"
)

type nodeslice []string

var nodes nodeslice
var staticRoot string

func (i *nodeslice) String() string {
	return fmt.Sprintf("%s", *i)
}

func (i *nodeslice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func connection(updates chan parser.SBabelUpdate, node string) {
	var conn net.Conn
	var err error

	for {
		log.Println("Trying", node)
		for {
			conn, err = net.Dial("tcp6", node)
			if err != nil {
				log.Println(err)
				time.Sleep(time.Second * 5)
			} else {
				break
			}
		}
		log.Println("Connected to", node)
		fmt.Fprintf(conn, "monitor\n")
		r := bufio.NewReader(conn)
		s := parser.NewScanner(r)
		desc := parser.NewBabelDesc()
		desc.Fill(s)
		ws.AddDesc(desc)
		err := desc.Listen(s, updates)
		if err != nil {
			log.Println(err)
			return
		}
		ws.RemoveDesc(desc.Id())
		conn.Close()
		log.Printf("Connection to %v closed\n", node)
	}
}

func main() {
	var bwPort string

	flag.Var(&nodes, "node",
		"Babel node to connect to (default \"[::1]:33123\", "+
			"may be repeated)")
	flag.StringVar(&bwPort, "http", ":8080", "web server address")
	flag.StringVar(&staticRoot, "static", "./static/",
		"directory with static files")
	flag.Parse()
	if len(nodes) == 0 {
		nodes = nodeslice{"[::1]:33123"}
	}

	ws.Init()

	updates := make(chan parser.SBabelUpdate, 1024)
	defer close(updates)

	for i := 0; i < len(nodes); i++ {
		go connection(updates, nodes[i])
	}

	bcastGrp := ws.NewListenerGroup()
	handler := ws.Handler(bcastGrp)
	http.Handle("/", http.FileServer(http.Dir(staticRoot)))
	http.Handle("/ws", handler)
	go func() {
		log.Fatal(http.ListenAndServe(bwPort, nil))
	}()

	for {
		upd := <-updates
		bcastGrp.Iter(func(l *ws.Listener) {
			l.Channel <- upd
		})
	}
}
