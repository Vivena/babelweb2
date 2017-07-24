package main

import (
	"babelweb2/parser"
	"babelweb2/ws"
	"bufio"
	"container/list"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"reflect"
	"sync"
	"time"
)

type connectslice []string

func (i *connectslice) String() string {
	return fmt.Sprintf("%s", *i)
}
func (i *connectslice) Set(value string) error {
	fmt.Printf("%s\n", value)
	*i = append(*i, value)
	return nil
}

var myconnectlist connectslice

func flagsInit(bwPort *string) int {

	flag.Var(&myconnectlist, "hp", "list of hostnames and portnums (shorthand)")
	flag.Var(&myconnectlist, "hostport", "liste of hostnames and portnums")

	flag.StringVar(bwPort, "b", ":8080", "babelweb Port (shorthand)")
	flag.StringVar(bwPort, "bwport", ":8080", "babelweb Port")
	flag.Parse()

	if flag.NFlag() == 0 {
		log.Println("connection to local node:")
	} else {
		fmt.Println("Here are the values")
		for i := 0; i < len(myconnectlist); i++ {
			fmt.Printf("-%s\n", myconnectlist[i])
		}
	}
	return flag.NFlag()
}

func connection(updates chan parser.BabelUpdate, wg *sync.WaitGroup, bwPort *string) {
	var node string
	node = "[::1]:33123"
	var lenArg = flagsInit(bwPort)
	log.Println("lenghth %d", lenArg)
	if lenArg == 0 {
		log.Println("to connect  to %s", node)
		go ConnectionNode(updates, node, wg)
	} else {
		go connectGroup(updates, node, wg)
	}
}

func ConnectionNode(updates chan parser.BabelUpdate, node string,
	wg *sync.WaitGroup) {
	var conn net.Conn
	var err error
	wg.Add(1)
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
			wg.Done()
			return
		}
		ws.Db.Lock()
		err = ws.Db.Bd.Clean(updates)
		ws.Db.Unlock()
		if err != nil {
			log.Println(err)
			wg.Done()
			return
		}
	}
	wg.Done()
}

var Listconduct = list.New()

func connectGroup(updates chan parser.BabelUpdate, node string, wg *sync.WaitGroup) {

	var wg2 sync.WaitGroup
	output := func(c chan parser.BabelUpdate) {
		for n := range c {
			updates <- n
		}
		//wg2.Done()
	}
	if Listconduct.Len() > 0 {
		wg2.Add(Listconduct.Len())
	}

	for i := 0; i < len(myconnectlist); i++ {
		conduct := make(chan parser.BabelUpdate, ws.ChanelSize)
		Listconduct.PushBack(conduct)
		go ConnectionNode(conduct, myconnectlist[i], wg)
	}

	cases := make([]reflect.SelectCase, Listconduct.Len())
	var i = 0
	for e := Listconduct.Front(); e != nil; e = e.Next() {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(e.Value)}
		i++
	}

	remaining := len(cases)
	for remaining > 0 {
		chosen, _, ok := reflect.Select(cases)
		if !ok {
			// The chosen channel has been closed
			cases[chosen].Chan = reflect.ValueOf(nil)
			remaining -= 1
			continue
		}
		if find(chosen) != nil {
			go output(find(chosen))
		} else {
			log.Println("null")
		}
	}
	wg2.Wait()
}

func find(index int) chan parser.BabelUpdate {
	var i = 0
	for e := Listconduct.Front(); e != nil; e = e.Next() {
		if i == (index) {
			return e.Value.(chan parser.BabelUpdate)
		} else {
			i++
		}
	}
	return nil
}

func main() {
	log.Println("	--------launching server--------")
	var bwPort string
	var wg sync.WaitGroup
	updates := make(chan parser.BabelUpdate, ws.ChanelSize)
	connection(updates, &wg, &bwPort)
	log.Println("valeur bxport%d", bwPort)

	ws.Db.Bd = parser.NewBabelDesc()
	bcastGrp := ws.NewListenerGroupe()
	go ws.MCUpdates(updates, bcastGrp, &wg)
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
