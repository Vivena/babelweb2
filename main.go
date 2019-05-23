package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Vivena/babelweb2/state"
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

func connection(updates chan state.Transition, node string,
	nl *ws.NodeList, delay time.Duration) {
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
		babel, err := state.NewBabelState(bufio.NewReader(conn), delay)
		if err != nil {
			// Don't you even dare to reconnect to this unholy node!
			log.Printf("Oh, boy! %v is doomed:\n\t%v.\t", node, err)
			afterHours()
			return
		} else {
			nl.Add(babel)
			err = babel.Listen(updates)
			if err != nil {
				log.Printf("Error while listening %v:\n\t%v.\n", node, err)
			}
			nl.Remove(babel)
		}
		afterHours()
	}
}

func main() {
	var (
		bwPort   string
		delayStr string
	)

	flag.Var(&nodes, "node",
		"Babel node to connect to (default \"[::1]:33123\", "+
			"may be repeated)")
	flag.StringVar(&bwPort, "http", ":8080", "Web server address")
	flag.StringVar(&staticRoot, "static", "static/",
		"Directory with static files")
	flag.StringVar(&delayStr, "delay", "0s", "Delay between updates")
	flag.Parse()

	delay, err := time.ParseDuration(delayStr)
	if err != nil {
		log.Fatalf("'%v': Not a duration.\n%v\n%v\n%v\n%v\n", delayStr,
			"A `delay` argument must be a string of possibly signed",
			"sequence of decimal numbers, each with optional fraction",
			"and a unit suffix. Valid time units are \"ns\", \"us\"",
			"(or \"µs\"), \"ms\", \"s\", \"m\", \"h\".")
	}

	if len(nodes) == 0 {
		nodes = nodeslice{"[::1]:33123"}
	}

	nl := ws.NewNodeList()
	updates := make(chan state.Transition, 1024)
	defer close(updates)

	for i := 0; i < len(nodes); i++ {
		go connection(updates, nodes[i], nl, delay)
	}

	_, err = os.Stat(staticRoot)
	if err != nil && os.IsNotExist(err) {
		log.Fatalf("'%v': No such directory\n%v\n%v.\n", staticRoot,
			"Try your best to find the directory containing the static files",
			"and precise it via the -static option")
	} else if err != nil {
		log.Fatalf("Something went terribly wrong: %v\n", err)
	}

	bcastGrp := ws.NewListenerGroup()
	handler := nl.Handler(bcastGrp)
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
