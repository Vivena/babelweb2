package ws

import (
	"container/list"
	"log"
	"sync"
)

const ChanelSize int = 1024

var globalClose = make(chan struct{})

//Listener unique channel for each ws
type Listener struct {
	conduct chan interface{}
	quit    chan struct{}
}

//WSMessage struc of messages sent by the ws to the clients

//Init create a Listener
func (l *Listener) Init() *Listener {
	l.conduct = make(chan interface{})
	l.quit = globalClose
	return l
}

//NewListener function to call if you want a new Listener
func NewListener() *Listener {
	return new(Listener).Init()
}

//Listenergroupe list of all the Listeners for the multicast
type Listenergroupe struct {
	sync.Mutex
	listeners *list.List
}

//Init create a new Listenergroupe
func (g *Listenergroupe) Init() *Listenergroupe {
	g.listeners = list.New()
	return g
}

//NewListenerGroupe function to call if you want a new Listenergroupe
func NewListenerGroupe() *Listenergroupe {
	return new(Listenergroupe).Init()
}

//Push add a Listener to the Listenergroupe
func (g *Listenergroupe) Push(newListener *Listener) {
	g.Lock()
	defer g.Unlock()

	g.listeners.PushBack(newListener)
}

//Flush remove a Listener from the Listenergroupe
func (g *Listenergroupe) Flush(l *Listener) {
	defer log.Println("Remouving listener from  the Listener groupe")
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
func (g *Listenergroupe) Iter(routine func(*Listener)) {
	g.Lock()
	defer g.Unlock()
	for i := g.listeners.Front(); i != nil; i = i.Next() {
		l := i.Value.(*Listener)
		routine(l)

	}
}
