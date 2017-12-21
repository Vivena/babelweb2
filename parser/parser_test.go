package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

// Simple test of lexical coherence.
// `TestListen` produces set of updates from the file and for each such
// update tests lexical equality with corresponding line of the source.
// By nature `TestListen` cannot be used as exhaustive parser test, but it's an
// important part of global testing.

func TestListen(t *testing.T) {
	r, err := os.Open("monitor")
	if err != nil {
		t.Fatal("os.Open:\n", err)
	}
	s := NewScanner(r)
	bd := NewBabelDesc()
	updChan := make(chan BabelUpdate)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = Listen(bd, s, updChan)
		wg.Done()
		close(updChan)
	}()
	r2, err := os.Open("monitor")
	if err != nil {
		t.Fatal("os.Open:\n", err)
	}
	reader := bufio.NewReader(r2)
	fieldStringHook := make(map[string](func(interface{}) string))
	fieldStringHook["neighbour_reach"] = func(v interface{}) string {
		return fmt.Sprintf("%04x", v)
	}
	fieldStringHook["route_installed"] = func(v interface{}) string {
		if b := v.(bool); b {
			return "yes"
		}
		return "no"
	}
	for upd := range updChan {
		upd.lequal(reader, t, fieldStringHook)
		if testing.Verbose() {
			fmt.Print(upd)
		}
	}
	wg.Wait()
	if err != nil {
		t.Fatal("parser.Listen:\n", err)
	}
}

// Lexical comparison of the given update with next valid line from `reader`.
// `fieldStringHook` associates entry field with 'toString-like' function.
// `lequal` uses `nextWord`, which must be tested separately.
// `lequal` is ugly and slow, but really simple
// (otherwise we need to test our test)

func (upd BabelUpdate) lequal(reader *bufio.Reader, t *testing.T,
	fieldStringHook map[string](func(interface{}) string)) {
	for {
		nextLine, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		r := strings.NewReader(nextLine)
		s := NewScanner(r)
		w, err := nextWord(s)
		if err != nil && err != io.EOF && err != errEOL {
			t.Fatal("parser.nextWord:\n", err)
		}
		if err == io.EOF || err == errEOL {
			return
		}
		line := w

		// In general, it's not an error,
		// but if it is, we need to be sure to track it.
		if w != string(upd.action) {
			fmt.Println("Warning: unknown action '" + w +
				"' (skip rest of the line)")
			for {
				w, err = nextWord(s)
				if err != nil && err != io.EOF &&
					err != errEOL {
					t.Fatal("parser.nextWord:\n", err)
				}
				if err == io.EOF || err == errEOL {
					if testing.Verbose() {
						fmt.Println(line)
					}
					break
				}
				line += (" " + w)
			}
			continue
		}
		checkNextWord := func(given string) {
			w, err = nextWord(s)
			if err != nil && err != io.EOF && err != errEOL {
				t.Fatal("parser.nextWord:\n", err)
			}
			if err == io.EOF || err == errEOL {
				if testing.Verbose() {
					fmt.Println(line)
				}
				return
			}
			line += (" " + w)
			if w != given {
				t.Fatal("In the line: ", line,
					"...\n\texpected: ", w,
					"\n\tgiven: ", given)
			}
		}
		checkNextWord(string(upd.tableId))
		checkNextWord(string(upd.entryId))
		for {
			w, err = nextWord(s)
			if err != nil && err != io.EOF && err != errEOL {
				t.Fatal("parser.nextWord:\n", err)
			}
			if err == io.EOF || err == errEOL {
				if testing.Verbose() {
					fmt.Println(line)
				}
				return
			}
			value, exists := upd.entry[Id(w)]
			if !exists {
				t.Fatal("No such field: " + w +
					"\n" + upd.entry.String())
			}
			line += (" " + w)
			w2, err := nextWord(s)
			if err != nil && err != io.EOF && err != errEOL {
				t.Fatal("parser.nextWord:\n", err)
			}
			if err == io.EOF || err == errEOL {
				if testing.Verbose() {
					fmt.Println(line)
				}
				return
			}
			line += (" " + w2)
			hook, exists := fieldStringHook[string(upd.tableId)+
				"_"+w]
			var given string
			if exists {
				given = hook(value.data)
			} else {
				given = fmt.Sprint(value.data)
			}
			if given != w2 {
				t.Fatal("In the line: ", line,
					"...\nexpected: ", w2,
					"\ngiven: ", given)
			}
		}
	}
}

func TestNextWord(t *testing.T) {
	input := "  Lorem ipsum dolor sit amet. Neighbour   55c47b990d90 " +
		"172.28.175.26/32-::/0\n" +
		"\"Now I have a machine gun. Ho-ho-ho.\"" + " \"\" " +
		"\"Who You Gonna Call?\nGhostbusters!\"" +
		" \"I have a bad feeling about t\"h\"is\" " +
		"\"Well, I called me wife and I said to her:\"" +
		"\"\\\"Will you kindly tell to me\n" +
		"Who owns that head upon the bed " +
		"where my old head should be?\\\"\"" +
		"\n\\\"A\" Spoonful of \"\"Sugar\"\\\""
	expect := []struct {
		word string
		err  error
	}{
		{"Lorem", nil},
		{"ipsum", nil},
		{"dolor", nil},
		{"sit", nil},
		{"amet.", nil},
		{"Neighbour", nil},
		{"55c47b990d90", nil},
		{"172.28.175.26/32-::/0", nil},
		{"", errEOL},
		{"Now I have a machine gun. Ho-ho-ho.", nil},
		{"", nil},
		{"Who You Gonna Call?\nGhostbusters!", nil},
		{"I have a bad feeling about this", nil},
		{"Well, I called me wife and I said to her:" +
			"\"Will you kindly tell to me\n" +
			"Who owns that head upon the bed " +
			"where my old head should be?\"", nil},
		{"", errEOL},
		{"\"A Spoonful of Sugar\"", nil},
		{"", io.EOF},
		{"", io.EOF},
	}
	r := strings.NewReader(input)
	s := NewScanner(r)
	for _, e := range expect {
		word, err := nextWord(s)
		if word != e.word || err != e.err {
			t.Errorf("nextWord:\nexpected: (%v, %v)\ngot: "+
				"(%v, %v)", e.word, e.err, word, err)
		}
	}
}

func Listen(bd *BabelDesc, s *Scanner, updChan chan BabelUpdate) error {
	for {
		upd, err := bd.ParseAction(s)
		if err != nil && err != io.EOF && err != errEOL {
			return err
		}
		if err == io.EOF {
			break
		}
		if upd.action != emptyUpdate.action {
			updChan <- upd
		}
	}
	return nil
}
