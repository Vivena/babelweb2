package ws

import (
	"container/list"
	"github.com/Vivena/babelweb2/parser"
	"log"
	"sync"
)

//Listener unique channel for each ws
type Listener struct {
	Channel chan parser.SBabelUpdate
}

//NewListener function to call if you want a new Listener
func NewListener() *Listener {
	l := new(Listener)
	l.Channel = make(chan parser.SBabelUpdate)
	return l
}

//Listenergroup list of all the Listeners for the multicast
type Listenergroup struct {
	sync.Mutex
	listeners *list.List
}

//NewListenerGroup function to call if you want a new Listenergroup
func NewListenerGroup() *Listenergroup {
	lg := new(Listenergroup)
	lg.listeners = list.New()
	return lg
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
