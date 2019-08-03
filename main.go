package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
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
		afterHours := func() {
			conn.Close()
			log.Printf("Connection to %v closed\n", node)
		}
		fmt.Fprintf(conn, "monitor\n")
		r := bufio.NewReader(conn)
		s := parser.NewScanner(r)
		desc := parser.NewBabelDesc()
		err = desc.Fill(s)
		if err == io.EOF {
			log.Printf("Something wrong with %v:\n\tcouldn't get router id.\n", node)
		} else if err != nil {
			// Don't you even dare to reconnect to this unholy node!
			log.Printf("Oh, boy! %v is doomed:\n\t%v.\t", node, err)
			afterHours()
			return
		} else {
			ws.AddDesc(desc)
			err = desc.Listen(s, updates)
			if err != nil {
				log.Printf("Error while listening %v:\n\t%v.\n", node, err)
			}
			ws.RemoveDesc(desc.ID)
		}
		afterHours()
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

	_, err := os.Stat(staticRoot)
	if err != nil && os.IsNotExist(err) {
		log.Fatalf("'%v': No such directory\n%v\n%v.\n", staticRoot,
			"Try your best to find the directory containing the static files",
			"and precise it via the -static option")
	} else if err != nil {
		log.Fatalf("Something went terribly wrong: %v\n", err)
	}

	bcastGrp := ws.NewListenerGroup()
	handler := ws.Handler(bcastGrp)
	http.Handle("/", http.FileServer(http.Dir(staticRoot)))
	http.Handle("/ws", handler)
	go func() {
		log.Printf("Listening on http://localhost%v\n", bwPort)
		log.Fatal(http.ListenAndServe(bwPort, nil))
	}()

	for {
		upd := <-updates
		bcastGrp.Iter(func(l *ws.Listener) {
			l.Channel <- upd
		})
	}
}
