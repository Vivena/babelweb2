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

var myconnectlist connectslice
var Listconduct = list.New()
var Quitmain = make(chan struct{}, 2)

func (i *connectslice) String() string {
	return fmt.Sprintf("%s", *i)
}
func (i *connectslice) Set(value string) error {
	fmt.Printf("%s\n", value)
	*i = append(*i, value)
	return nil
}

func flagsInit(bwPort *string) int {

	flag.Var(&myconnectlist, "hp", "list of hostnames and portnums (shorthand)")
	flag.Var(&myconnectlist, "hostport", "liste of hostnames and portnums")

	flag.StringVar(bwPort, "b", ":8080", "babelweb Port (shorthand)")
	flag.StringVar(bwPort, "bwport", ":8080", "babelweb Port")
	flag.Parse()

	if len(myconnectlist) == 0 {
		log.Println("connection to local node:")
	} else {
		fmt.Println("Here are the values")
		for i := 0; i < len(myconnectlist); i++ {
			fmt.Printf("-%s\n", myconnectlist[i])
		}
	}
	return len(myconnectlist)
}

func connection(updates chan parser.BabelUpdate, wg *sync.WaitGroup, bwPort *string) {
	var node string
	node = "[::1]:33123"
	var lenArg = flagsInit(bwPort)
	log.Println("lenghth %d", lenArg)
	if lenArg == 0 {
		log.Println("to connect  to %s", node)
		go ConnectionNode(updates, node, wg, Quitmain)
	} else {
		go connectGroup(updates, node, wg)
	}
}

func ConnectionNode(updates chan parser.BabelUpdate, node string,
	wg *sync.WaitGroup, quit chan struct{}) {
	var conn net.Conn
	var err error
	exit := true
	wg.Add(1)
	defer wg.Done()
	defer close(updates)

	for {
		select {
		case _, q := <-quit:
			if !q {
				return
			}
		default:
			log.Println("	Trying ", node)
			for exit {
				select {
				case _, q := <-quit:
					if !q {
						return
					}
				default:
					conn, err = net.Dial("tcp6", node)
					if err != nil {
						log.Println(err)
						time.Sleep(time.Second * 5)
					} else {
						log.Println("test")
						exit = false
					}
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
				return
			}
			ws.Db.Lock()
			err = ws.Db.Bd.Clean(updates)
			ws.Db.Unlock()
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func connectGroup(updates chan parser.BabelUpdate, node string, wg *sync.WaitGroup) {
	var quitgroup = make(chan struct{}, 2)
	var wgGroup sync.WaitGroup

	wg.Add(1)
	defer close(updates)
	defer wg.Done()
	defer wgGroup.Wait()
	defer close(quitgroup)

	output := func(c chan parser.BabelUpdate) {
		for n := range c {
			updates <- n
		}
	}

	for i := 0; i < len(myconnectlist); i++ {
		conduct := make(chan parser.BabelUpdate, ws.ChanelSize)
		Listconduct.PushBack(conduct)
		go ConnectionNode(conduct, myconnectlist[i], &wgGroup, quitgroup)
	}

	cases := make([]reflect.SelectCase, Listconduct.Len())
	var i = 0
	for e := Listconduct.Front(); e != nil; e = e.Next() {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(e.Value)}
		i++
	}

	remaining := len(cases)
	for remaining > 0 {
		select {
		case _, q := <-Quitmain:
			if !q {
				return
			}
		default:
			chosen, _, ok := reflect.Select(cases)
			if !ok {
				cases[chosen].Chan = reflect.ValueOf(nil)
				remaining -= 1
				continue
			}
			if find(chosen) != nil {
				output(find(chosen))
			} else {
				log.Println("null")
			}
		}
	}
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

	defer close(Quitmain)
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
