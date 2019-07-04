package state

import (
	"io"
	"sync"
	"time"

	"github.com/Vivena/babelweb2/parser"
)

type TransitionKey struct {
	table string
	field string
}

type TransitionSet struct {
	sync.Mutex
	ts map[TransitionKey]*parser.Transition
}

type BabelState struct {
	parser  *parser.Parser
	delay   time.Duration
	history *TransitionSet
	news    *TransitionSet
}

const DEFAULT_SET_SIZE = 100

func newTransitonSet() *TransitionSet {
	return &TransitionSet{
		ts: make(map[TransitionKey]*parser.Transition, DEFAULT_SET_SIZE),
	}
}

func (set *TransitionSet) add(t *parser.Transition) {
	if t == nil {
		return
	}

	key := TransitionKey{
		table: t.Table,
		field: t.Field,
	}

	set.Lock()
	defer set.Unlock()

	if _, exists := set.ts[key]; exists && t.Action == "flush" {
		delete(set.ts, key)
	} else {
		set.ts[key] = t
	}
}

func NewBabelState(reader io.Reader, delay time.Duration) (*BabelState, error) {
	state := &BabelState{
		parser:  parser.NewParser(reader),
		delay:   delay,
		history: newTransitonSet(),
		news:    newTransitonSet(),
	}
	err := state.parser.Init()

	if err != nil {
		state = nil
	}
	return state, err
}

func (b *BabelState) Version() string {
	return b.parser.Version
}

func (b *BabelState) Iter(f func(t parser.Transition) error) error {
	b.history.Lock()
	defer b.history.Unlock()

	for _, t := range b.history.ts {
		err := f(*t)
		if err != nil {
			return nil
		}
	}

	return nil
}

func (b *BabelState) Listen(updates chan parser.Transition) error {
	defer func() {
		ts := make(map[*parser.Transition]struct{}, len(b.history.ts))

		b.history.Lock()
		for _, t := range b.history.ts {
			t.Action = "flush"
			ts[t] = struct{}{}
		}
		b.history.Unlock()

		for t := range ts {
			b.history.add(t)
			b.news.add(t)
			updates <- *t
		}
	}()

	active := true

	if b.delay != time.Duration(0) {
		go func() {
			for active {
				time.Sleep(b.delay)
				b.news.Lock()
				for _, t := range b.news.ts {
					updates <- *t
				}
				b.news.ts = make(map[TransitionKey]*parser.Transition,
					DEFAULT_SET_SIZE)
				b.news.Unlock()
			}
		}()
	}

	for {
		t, err := b.parser.Parse()
		if err != nil {
			active = false
			return err
		}
		if t == nil {
			continue
		}

		st := parser.Transition(*t)

		b.history.add(&st)
		if b.delay != time.Duration(0) {
			b.news.add(&st)
		} else {
			updates <- st
		}
	}

	active = false
	return nil
}
