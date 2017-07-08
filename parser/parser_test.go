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
	br := bufio.NewReader(r)
	bd := NewBabelDesc()
	err = bd.Fill(br)
	if err != nil {
		t.Error(err)
	}
	if testing.Verbose() {
		fmt.Println("Fill\n", bd)
	}
	err = bd.Fill(br)
	if err != nil {
		t.Error(err)
	}
	if testing.Verbose() {
		fmt.Println("Update\n", bd)
		fmt.Println(bd)
	}
}
