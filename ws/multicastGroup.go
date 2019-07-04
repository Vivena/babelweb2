package ws

import (
	"container/list"
	"sync"

	"github.com/Vivena/babelweb2/parser"
)

type Listener struct {
	Channel chan parser.Transition
}

type Listenergroup struct {
	sync.Mutex
	listeners *list.List
}

func NewListener() *Listener {
	l := new(Listener)
	l.Channel = make(chan parser.Transition)
	return l
}

func NewListenerGroup() *Listenergroup {
	lg := new(Listenergroup)
	lg.listeners = list.New()
	return lg
}

func (g *Listenergroup) Push(newListener *Listener) {
	g.Lock()
	defer g.Unlock()

	g.listeners.PushBack(newListener)
}

func (g *Listenergroup) Flush(l *Listener) {
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

func (g *Listenergroup) Iter(routine func(*Listener)) {
	g.Lock()
	defer g.Unlock()

	for i := g.listeners.Front(); i != nil; i = i.Next() {
		l := i.Value.(*Listener)
		routine(l)
	}
}
