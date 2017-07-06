package main

import (
	"container/list"
	"log"
	"sync"
)

const chanelSize int = 1024

var globalClose = make(chan struct{})

type Listener struct {
	conduct chan interface{}
	quit    chan struct{}
}

func (l *Listener) Init() *Listener {
	l.conduct = make(chan interface{})
	l.quit = globalClose
	return l
}

func NewListener() *Listener {
	return new(Listener).Init()
}

type Listenergroupe struct {
	sync.Mutex
	listeners *list.List
}

func (l *Listenergroupe) Init() *Listenergroupe {
	l.listeners = list.New()
	return l
}

func NewListenerGroupe() *Listenergroupe {
	return new(Listenergroupe).Init()
}

func (g *Listenergroupe) Push(newListener *Listener) {
	g.Lock()
	defer g.Unlock()

	g.listeners.PushBack(newListener)
}

func (g *Listenergroupe) Quit(l *Listener) {
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

func (g *Listenergroupe) Iter(routine func(*Listener)) {
	g.Lock()
	defer g.Unlock()
	for i := g.listeners.Front(); i != nil; i = i.Next() {
		l := i.Value.(*Listener)
		routine(l)

	}
}
