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

const (
	dump    = "dump\n"
	monitor = "monitor\n"
	node    = "[::1]:33123"
)

var Bd parser.BabelDesc

func Connection(updates chan parser.BabelUpdate, node string) {
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
	updates := make(chan parser.BabelUpdate, ws.ChanelSize)
	Bd = parser.NewBabelDesc()
	log.Println("test1")
	go Connection(updates, node)
	log.Println("test2")
	go ws.Manager(updates)

	wg.Wait()
}
