package ws

import (
	"container/list"
	"github.com/Vivena/babelweb2/parser"
	"log"
	"sync"
)

const ChanelSize int = 1024

var globalClose = make(chan struct{})

//Listener unique channel for each ws
type Listener struct {
	conduct chan parser.SBabelUpdate
}

//Init create a Listener
func (l *Listener) Init() *Listener {
	l.conduct = make(chan parser.SBabelUpdate)
	return l
}

//NewListener function to call if you want a new Listener
func NewListener() *Listener {
	return new(Listener).Init()
}

//Listenergroup list of all the Listeners for the multicast
type Listenergroup struct {
	sync.Mutex
	listeners *list.List
}

//Init create a new Listenergroup
func (g *Listenergroup) Init() *Listenergroup {
	g.listeners = list.New()
	return g
}

//NewListenerGroup function to call if you want a new Listenergroup
func NewListenerGroup() *Listenergroup {
	return new(Listenergroup).Init()
}

//Push add a Listener to the Listenergroup
func (g *Listenergroup) Push(newListener *Listener) {
	g.Lock()
	defer g.Unlock()

	g.listeners.PushBack(newListener)
}

//Flush remove a Listener from the Listenergroup
func (g *Listenergroup) Flush(l *Listener) {
	defer log.Println("Remouving listener from  the Listener group")
	g.Lock()
	defer g.Unlock()

	for i := g.listeners.Front(); i != nil; i = i.Next() {
		current := i.Value.(*Listener)
		if current == l {
			g.listeners.Remove(i)
			return
		}
	}
}

//Iter Call the routine for each Listener
func (g *Listenergroup) Iter(routine func(*Listener)) {
	g.Lock()
	defer g.Unlock()
	for i := g.listeners.Front(); i != nil; i = i.Next() {
		l := i.Value.(*Listener)
		routine(l)

	}
}
