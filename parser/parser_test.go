package parser

import (
	"bufio"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	r := strings.NewReader(
		"add interface enp3s0 up true "         +
		"ipv6 fe80::3649:9d87:1d91:ce03\n"      +
		"add neighbour 55c47b990d90 address "   +
		"fe80::e046:9aff:fe4e:912e if enp3s0 "  +
		"reach fff0 rxcost 96 txcost 96 cost 96")
	br := bufio.NewReader(r)
	bd := NewBabelDesc()
	err := bd.Fill(br)
	if err != nil {
		t.Error(err)
	}
}
