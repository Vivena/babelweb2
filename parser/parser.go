package parser

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type sReader struct {
	ligne []string
	index int
}

func (b *sReader)nextLine(reader *bufio.Reader) error {
	w, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return err
	}
	b.ligne = strings.Fields(string(w))
	b.index = 0
	return err
}

func (b *sReader)nextWord() (string, error) {
	if b.index == len(b.ligne) {
		return "", io.EOF
	}
	defer func (){
		b.index++
	}()
	return b.ligne[b.index], nil
}

type Id string

type EntryParser func(*sReader) (interface{}, error)

type EntryValue struct {
	data interface{}
	parser EntryParser
}

type Entry map[Id](*EntryValue)

func (e Entry)String() string {
	var s string
	for id, ev := range e {
		s += (fmt.Sprintf("\t%s: ", id) +
			fmt.Sprintln(ev.data))
	}
	return s
}

type EntryError int

const (
	FieldPresence EntryError = 0
	FieldAbsence  EntryError = 1
)

func (e EntryError)Error() string {
	if e == FieldPresence {
		return "Field Already Exists"
	} else if e == FieldAbsence {
		return "No Such Field"
	}
	return "Error of Lack of Error"
}

func NewEntry() Entry {
	return make(map[Id](*EntryValue))
}

func (e *Entry)AddField(id Id, parser EntryParser) error {
	_, exists := (*e)[id]
	if exists {
		return FieldPresence
	}
	(*e)[id] = new(EntryValue)
	(*e)[id].data = nil
	(*e)[id].parser = parser
	return nil
}

func (e *Entry)GetData(id Id) (interface{}, error) {
	value, exists := (*e)[id]
	if !exists {
		return nil, FieldAbsence
	}
	return value.data, nil
}

func (e *Entry)Parse(buf *sReader) error {
	for {
		w, err := buf.nextWord()
		if err != nil {
			return err
		}
		value, exists := (*e)[Id(w)]
		if !exists {
			_, err = buf.nextWord()
			if err != nil {
				return err
			}
			continue
		}
		new_data, err := value.parser(buf)
		if err != nil && err != io.EOF {
			return err
		}
		value.data = new_data
		if err == io.EOF {
			return err
		}
	}
	return nil
}

func NewInterfaceEntry() Entry {
	i := NewEntry()
	i.AddField("up", ParseBool)
	i.AddField("ipv4", ParseIp)
	i.AddField("ipv6", ParseIp)
	return i
}

func NewNeighbourEntry() Entry {
	i := NewEntry()
	i.AddField("address", ParseIp)
	i.AddField("if", ParseString)
	i.AddField("reach", GetUintParser(16, 16))
	i.AddField("rxcost", GetUintParser(10, 16))
	i.AddField("txcost", GetUintParser(10, 16))
	i.AddField("cost", GetUintParser(10, 16))
	i.AddField("rtt", ParseString)
	i.AddField("rttcost", GetUintParser(10, 16))
	return i
}

func NewRouteEntry() Entry {
	i := NewEntry()
	i.AddField("prefix", ParsePrefix)
	i.AddField("from", ParsePrefix)
	i.AddField("installed", ParseBool)
	i.AddField("id", ParseString)
	i.AddField("metric", GetIntParser(10, 16))
	i.AddField("refmetric", GetUintParser(10, 16))
	i.AddField("via", ParseIp)
	i.AddField("if", ParseString)
	return i
}

func NewXrouteEntry() Entry {
	i := NewEntry()
	i.AddField("prefix", ParsePrefix)
	i.AddField("from", ParsePrefix)
	i.AddField("meric", GetUintParser(10, 16))
	return i
}

type ParsErr int

func (e ParsErr)Error() string {
	return "Syntax Error"
}

const SyntaxError ParsErr = 0

// string
func ParseString(buf *sReader) (interface{}, error) {
	return buf.nextWord()
}

// bool
func ParseBool(buf *sReader) (interface{}, error) {
	w, err := buf.nextWord()
	if err != nil {
		return nil, err
	}
	if w == "true" || w == "yes" || w == "oui" ||
		w == "tak" || w == "да" {
		return true, nil
	} else if w == "false" || w == "no" || w == "non" ||
		w == "nie" || w == "нет" {
		return false, nil
	}
	return nil, SyntaxError
}

// int64
func GetIntParser(base int, bitSize int) EntryParser {
	return func(buf *sReader) (interface{}, error) {
		w, err := buf.nextWord()
		if err != nil {
			return nil, err
		}
		i, err := strconv.ParseInt(w, base, bitSize)
		if err != nil {
			return nil, err
		}
		return i, nil
	}
}

// uint64
func GetUintParser(base int, bitSize int) EntryParser {
	return func(buf *sReader) (interface{}, error) {
		w, err := buf.nextWord()
		if err != nil {
			return nil, err
		}
		i, err := strconv.ParseUint(w, base, bitSize)
		if err != nil {
			return nil, err
		}
		return i, nil
	}
}

// net.IP
func ParseIp(buf *sReader) (interface{}, error) {
	w, err := buf.nextWord()
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(w)
	if ip == nil {
		return nil, SyntaxError
	}
	return ip, nil
}

// *net.IPNet
func ParsePrefix(buf *sReader) (interface{}, error) {
	w, err := buf.nextWord()
	if err != nil {
		return nil, err
	}
	_, ip, err := net.ParseCIDR(w)
	if err != nil {
		return nil, SyntaxError
	}
	return ip, nil
}

type EntryMaker func() Entry

type Table struct {
	dict map[Id](Entry)
	maker EntryMaker
}

func (t Table)String() string {
	var s string
	for id, e := range t.dict {
		s += (fmt.Sprintf("%s:\n", id) +
			fmt.Sprintln(e))
	}
	return s
}

type BabelDesc map[Id](Table)

func (bd BabelDesc)String() string {
	var s string
	for id, t := range bd {
		s += (fmt.Sprintf("*\t%s\n", id) +
			fmt.Sprintln(t))
	}
	return s
}

func NewBabelDesc() BabelDesc {
	ts := make(map[Id](Table))
	ts["route"] = Table{make(map[Id](Entry)), NewRouteEntry}
	ts["xroute"] = Table{make(map[Id](Entry)), NewXrouteEntry}
	ts["interface"] = Table{make(map[Id](Entry)), NewInterfaceEntry}
	ts["neighbour"] = Table{make(map[Id](Entry)), NewNeighbourEntry}
	return ts
}

func (t Table)Add(id Id, e Entry) error {
	_, exists := t.dict[id]
	if exists {
		return FieldPresence
	}
	t.dict[id] = e
	return nil
}

func (t Table)Change(id Id, e Entry) error {
	_, exists := t.dict[id]
	if !exists {
		return FieldAbsence
	}
	t.dict[id] = e
	return nil
}

func (t Table)Flush(id Id) error {
	_, exists := t.dict[id]
	if !exists {
		return FieldAbsence
	}
	delete(t.dict, id)
	return nil
}

func ParseAction(t *BabelDesc, buf *sReader) error {
	w, err := buf.nextWord()
	if err != nil {
		return err
	}
	if w != "add" && w != "change" && w != "flush" {
		return nil
	}
	table_id, err := buf.nextWord()
	if err != nil {
		return err
	}
	entry_id, err := buf.nextWord()
	if err != nil {
		return err
	}
	new_entry := (*t)[Id(table_id)].maker()
	err = new_entry.Parse(buf)
	if err != io.EOF {
		return err
	}
	switch w {
	case "add":
		return (*t)[Id(table_id)].Add(Id(entry_id), new_entry)
	case "change":
		return (*t)[Id(table_id)].Change(Id(entry_id), new_entry)
	case "flush":
		return (*t)[Id(table_id)].Flush(Id(entry_id))
	}
	return nil
}

func (t *BabelDesc)Fill(reader *bufio.Reader) error {
	var buf sReader
	for {
		err := buf.nextLine(reader)
		if err != nil && err != io.EOF {
			return err
		}
		parserr := ParseAction(t, &buf)
		if parserr != nil {
			return parserr
		}
		if err == io.EOF {
			break
		}
	}
	return nil
}
