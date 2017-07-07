package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/babelweb2/parser"
	"github.com/babelweb2/ws"
)

//const node string = "[fe80::e046:9aff:fe4e:912e%wlp1s0]:33123"

//const node string = "[fe80::e046:9aff:fe4e:912e%enp2s0]:33123"

//const node string = "[fe80::1e8f:814e:9731:dec6%enp2s0]:33123"

const (
	dump = "dump\n"
	monitor = "monitor\n"
	node = "[::1]:33123"
)

var bd parser.BabelDesc

func Connection(updates chan interface{}, node string) {
	conn, err := net.Dial("tcp6", node)
	if err != nil {
		return
	}
	defer conn.Close()
	fmt.Fprintf(conn, monitor)
	bd := parser.NewBabelDesc()
	r := bufio.NewReader(conn)
	for {
		bd.Fill(r)
	}
}

/*
func testConnection() {
	conn, err := net.Dial("tcp6", node)
	if err != nil {
		log.Println("node ", err)
		return
	}
	defer conn.Close()
	fmt.Fprintf(conn, dump)
	bd := NewBabelDesc()
	
	
	
	/*
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println("no")
			break
		}
		log.Println(message)
		if message == "ok" || message == "bad" || message == "no" {
			break
		}
	}*/
//}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	updates := make(chan interface{}, ws.ChanelSize)
	log.Println("test1")
	//go testConnection()
	go Connection(updates, node)
	log.Println("test2")
	go ws.WsManager(updates)

	wg.Wait()
}
