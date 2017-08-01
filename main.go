package main

import (
	"bufio"
	"container/list"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
	"github.com/Vivena/babelweb2/parser"
	"github.com/Vivena/babelweb2/ws"
)

type connectslice []string

var myconnectlist connectslice
var Listconduct = list.New()
var Quitmain = make(chan struct{}, 2)
var static_root string

func (i *connectslice) String() string {
	return fmt.Sprintf("%s", *i)
}
func (i *connectslice) Set(value string) error {
	fmt.Printf("%s\n", value)
	*i = append(*i, value)
	return nil
}

func flagsInit(bwPort *string) int {

	flag.Var(&myconnectlist, "hp",
		"list of hostnames and portnums (shorthand)")
	flag.Var(&myconnectlist, "hostport", "liste of hostnames and portnums")

	flag.StringVar(bwPort, "b", ":8080", "babelweb Port (shorthand)")
	flag.StringVar(bwPort, "bwport", ":8080", "babelweb Port")
	flag.StringVar(&static_root, "static", "./static/",
		"directory with static files")
	flag.Parse()

	return len(myconnectlist)
}

func connection(updates chan parser.BabelUpdate,
	wg *sync.WaitGroup, bwPort *string) {
	var node string
	node = "[::1]:33123"
	var lenArg = flagsInit(bwPort)
	if lenArg == 0 {
		wg.Add(1)
		go func () {
			ConnectionNode(updates, node, Quitmain)
			wg.Done()
		}()
	} else {
		connectGroup(updates, wg)
	}
}

func ConnectionNode(updates chan parser.BabelUpdate,
	node string, quit chan struct{}) {
	var conn net.Conn
	var err error

	for {
		exit := true
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
						exit = false
					}
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
}

func connectGroup(updates chan parser.BabelUpdate, wg *sync.WaitGroup) {
	var quitgroup = make(chan struct{}, 2)

	for i := 0; i < len(myconnectlist); i++ {
		wg.Add(1)
		go func() {
			ConnectionNode(updates, myconnectlist[i], quitgroup)
			wg.Done()
		}()
	}
}

func main() {
	ws.Init()
	defer close(Quitmain)
	log.Println("	--------launching server--------")
	var bwPort string
	var wg sync.WaitGroup

	updates := make(chan parser.BabelUpdate, ws.ChanelSize)
	defer close(updates)

	connection(updates, &wg, &bwPort)
	bcastGrp := ws.NewListenerGroup()
	wg.Add(1)
	go func () {
		ws.MCUpdates(updates, bcastGrp)
		wg.Done()
	}()
	ws := ws.Handler(bcastGrp)

	static := static_root
	if !strings.HasSuffix(static, "/") {
		static = static + "/"
	}

	http.Handle("/", http.FileServer(http.Dir(static)))
	http.Handle("/style.css", http.FileServer(http.Dir(static + "css/")))
	http.Handle("/initialize.js", http.FileServer(http.Dir(static + "js/")))
	http.Handle("/d3/d3.js", http.FileServer(http.Dir(static + "js/")))
	http.Handle("/ws", ws)

	err := http.ListenAndServe(bwPort, nil)
	if err != nil {
		log.Println(err)
		return
	}
	wg.Wait()
}
