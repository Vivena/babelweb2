package main

import (
	"babelweb2/parser"
	"babelweb2/ws"
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

//const node string = "[fe80::e046:9aff:fe4e:912e%wlp1s0]:33123"

//const node string = "[fe80::e046:9aff:fe4e:912e%enp2s0]:33123"

//const node string = "[fe80::1e8f:814e:9731:dec6%enp2s0]:33123"

const (
	dump    = "dump\n"
	monitor = "monitor\n"
	node    = "[::1]:33123"
)

var Bd parser.BabelDesc

func Connection(updates chan interface{}, node string) {
	conn, err := net.Dial("tcp6", node)
	if err != nil {
		log.Println("node ", err)
		return
	}
	defer conn.Close()
	fmt.Fprintf(conn, monitor)
	r := bufio.NewReader(conn)
	s := bufio.NewScanner(r)
	for {
		Bd.Listen(s, updates)
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	updates := make(chan interface{}, ws.ChanelSize)
	Bd = parser.NewBabelDesc()
	log.Println("test1")
	go Connection(updates, node)
	log.Println("test2")
	go ws.Manager(updates)

	wg.Wait()
}
