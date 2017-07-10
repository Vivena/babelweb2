package parser

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

func TestParser(t *testing.T) {
	r, err := os.Open("monitor")
	if err != nil {
		t.Error(err)
	}
	s := bufio.NewScanner(r)
	bd := NewBabelDesc()
	updChan := make(chan BabelUpdate)
	go bd.Listen(s, updChan)
	for upd := range updChan {
		if testing.Verbose() {
			fmt.Print(upd)
		}
		err = bd.Update(upd)
		if err != nil {
			t.Error(err)
		}
	}
	if testing.Verbose() {
		fmt.Printf("\n%s\n", bd)
	}
}
