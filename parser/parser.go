package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"strconv"
)

var errEOL = errors.New("EOL")
var ErrUnterminatedString = errors.New("Unterminated String")

type Scanner struct {
	bufio.Scanner
}

func nextWord(s *Scanner) (string, error) {
	more := s.Scan()
	if more {
		if s.Text() == "\n" {
			return "", errEOL
		} else {
			return s.Text(), nil
		}
	}
	err := s.Err()
	if err == nil {
		return "", io.EOF
	}
	return "", err
}

type Id string

type EntryParser func(*Scanner) (interface{}, error)

type EntryValue struct {
	data   interface{}
	parser EntryParser
}

type Entry map[Id](*EntryValue)

func (e Entry) String() string {
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

func (e EntryError) Error() string {
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

func (e *Entry) AddField(id Id, parser EntryParser) error {
	_, exists := (*e)[id]
	if exists {
		return FieldPresence
	}
	(*e)[id] = new(EntryValue)
	(*e)[id].data = nil
	(*e)[id].parser = parser
	return nil
}

func (e *Entry) GetData(id Id) (interface{}, error) {
	value, exists := (*e)[id]
	if !exists {
		return nil, FieldAbsence
	}
	return value.data, nil
}

func (e *Entry) Parse(s *Scanner) error {
	for {
		w, err := nextWord(s)
		if err != nil {
			return err
		}
		value, exists := (*e)[Id(w)]
		if !exists {
			_, err = nextWord(s)
			if err != nil {
				return err
			}
			continue
		}
		new_data, err := value.parser(s)
		if err != nil && err != io.EOF && err != errEOL {
			return err
		}
		value.data = new_data
		if err == io.EOF || err == errEOL {
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
	i.AddField("rxcost", GetUintParser(10, 32))
	i.AddField("txcost", GetUintParser(10, 32))
	i.AddField("cost", GetUintParser(10, 32))
	i.AddField("rtt", ParseString)
	i.AddField("rttcost", GetUintParser(10, 32))
	return i
}

func NewRouteEntry() Entry {
	i := NewEntry()
	i.AddField("prefix", ParsePrefix)
	i.AddField("from", ParsePrefix)
	i.AddField("installed", ParseBool)
	i.AddField("id", ParseString)
	i.AddField("metric", GetIntParser(10, 32))
	i.AddField("refmetric", GetUintParser(10, 32))
	i.AddField("via", ParseIp)
	i.AddField("if", ParseString)
	return i
}

func NewXrouteEntry() Entry {
	i := NewEntry()
	i.AddField("prefix", ParsePrefix)
	i.AddField("from", ParsePrefix)
	i.AddField("metric", GetUintParser(10, 32))
	return i
}

// string
func ParseString(s *Scanner) (interface{}, error) {
	return nextWord(s)
}

// bool
func ParseBool(s *Scanner) (interface{}, error) {
	w, err := nextWord(s)
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
	return nil, errors.New("Syntax Error: '" + w + "' must be a boolean")
}

// int64
func GetIntParser(base int, bitSize int) EntryParser {
	return func(s *Scanner) (interface{}, error) {
		w, err := nextWord(s)
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
	return func(s *Scanner) (interface{}, error) {
		w, err := nextWord(s)
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
func ParseIp(s *Scanner) (interface{}, error) {
	w, err := nextWord(s)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(w)
	if ip == nil {
		return nil, errors.New("Syntax Error: invalid IP address: " + w)
	}
	return ip, nil
}

// *net.IPNet
func ParsePrefix(s *Scanner) (interface{}, error) {
	w, err := nextWord(s)
	if err != nil {
		return nil, err
	}
	_, ip, err := net.ParseCIDR(w)
	if err != nil {
		return nil, errors.New("Syntax Error: " + err.Error())
	}
	return ip, nil
}

type EntryMaker func() Entry

type Table struct {
	dict  map[Id](Entry)
	maker EntryMaker
}

func (t Table) String() string {
	var s string
	for id, e := range t.dict {
		s += (fmt.Sprintf("%s:\n", id) +
			fmt.Sprintln(e))
	}
	return s
}

type BabelDesc map[Id](Table)

func (bd BabelDesc) String() string {
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

func (t Table) Add(id Id, e Entry) error {
	_, exists := t.dict[id]
	if exists {
		return FieldPresence
	}
	t.dict[id] = e
	return nil
}

func (t Table) Change(id Id, e Entry) error {
	_, exists := t.dict[id]
	if !exists {
		return FieldAbsence
	}
	t.dict[id] = e
	return nil
}

func (t Table) Flush(id Id) error {
	_, exists := t.dict[id]
	if !exists {
		return FieldAbsence
	}
	delete(t.dict, id)
	return nil
}

type BabelUpdate struct {
	action  Id
	tableId Id
	entryId Id
	entry   Entry
}

type SBabelUpdate struct {
	Action    Id                 `json:"action"`
	TableId   Id                 `json:"table"`
	EntryId   Id                 `json:"id"`
	EntryData map[Id]interface{} `json:"data"`
}

func (bd BabelDesc) Iter(f func(BabelUpdate) error) error {
	for tk, tv := range bd {
		for ek, ev := range tv.dict {
			err := f(BabelUpdate{action: "add", tableId: tk,
				entryId: ek, entry: ev})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (upd BabelUpdate) ToS() SBabelUpdate {
	s_upd := SBabelUpdate{upd.action, upd.tableId, upd.entryId,
		make(map[Id]interface{})}
	for id, ev := range upd.entry {
		s_upd.EntryData[id] = ev.data
	}
	return s_upd
}

var emptyUpdate = BabelUpdate{action: Id("none")}

func (upd BabelUpdate) String() string {
	return fmt.Sprintf("%s: %s %s\n%s", upd.action, upd.tableId,
		upd.entryId, upd.entry)

}

func (bd *BabelDesc) ParseAction(s *Scanner) (BabelUpdate, error) {
	w, err := nextWord(s)
	if err != nil {
		return emptyUpdate, err
	}
	if w != "add" && w != "change" && w != "flush" {
		return emptyUpdate, nil
	}
	table_id, err := nextWord(s)
	if err != nil {
		return emptyUpdate, err
	}
	entry_id, err := nextWord(s)
	if err != nil {
		return emptyUpdate, err
	}
	new_entry := (*bd)[Id(table_id)].maker()
	err = new_entry.Parse(s)
	if err != io.EOF && err != errEOL {
		return emptyUpdate, err
	}
	return BabelUpdate{Id(w), Id(table_id), Id(entry_id), new_entry}, err
}

func (bd *BabelDesc) Update(upd BabelUpdate) error {
	switch upd.action {
	case "add":
		return (*bd)[Id(upd.tableId)].Add(
			Id(upd.entryId), upd.entry)
	case "change":
		return (*bd)[Id(upd.tableId)].Change(
			Id(upd.entryId), upd.entry)
	case "flush":
		return (*bd)[Id(upd.tableId)].Flush(Id(upd.entryId))
	}
	return nil
}

func (bd *BabelDesc) CheckUpdate(upd BabelUpdate) bool {
	if upd.action != Id("change") {
		return true
	}
	for key, value := range (*bd)[Id(upd.tableId)].dict[Id(upd.entryId)] {
		if !(reflect.DeepEqual((*upd.entry[key]).data, (*value).data)) {
			return true
		}
	}
	return false
}

func split(data []byte, atEOF bool) (int, []byte, error) {
	start := 0
	for start < len(data) && (data[start] == ' ' || data[start] == '\r') {
		start++
	}

	if start < len(data) && data[start] == '\n' {
		return start + 1, []byte{'\n'}, nil
	}

	split_quotes := func(start int) (int, []byte, error) {
		start++
		i := start
		token := ""
		b := false
		for i < len(data) && data[i] != '"' {
			if i < len(data)-1 && data[i] == '\\' &&
				data[i+1] == '"' {
				token += (string(data[start:i]) + "\"")
				i += 2
				start = i
				b = true
			} else {
				i++
				b = false
			}
		}

		if i < len(data) {
			if b {
				return i + 1, []byte(token), nil
			}
			token += string(data[start:i])
			return i + 1, []byte(token), nil
		}
		return 0, nil, ErrUnterminatedString
	}
	i := start
	token := ""
	b := false
	for i < len(data) && data[i] != ' ' && data[i] != '\r' &&
		data[i] != '\n' {
		if i < len(data)-1 && data[i] == '\\' && data[i+1] == '"' {
			token += "\""
			i += 2
			start = i
		} else if i < len(data) && data[i] == '"' {
			token += string(data[start:i])
			n, quotok, err := split_quotes(i)
			b = true
			if err != nil {
				return n, quotok, err
			}
			token += string(quotok)
			i = n
			start = i
		} else {
			i++
			b = false
		}
	}

	if b {
		return i, []byte(token), nil
	}
	if i < len(data) {
		token += string(data[start:i])
		return i, []byte(token), nil
	}
	if atEOF && start < len(data) {
		token += string(data[start:])
		return len(data), []byte(token), nil
	}

	return start, nil, nil
}

func NewScanner(r io.Reader) *Scanner {
	s := bufio.NewScanner(r)
	s.Split(split)
	return &Scanner{*s}
}

func (t *BabelDesc) Listen(s *Scanner, updChan chan BabelUpdate) error {
	for {
		upd, err := t.ParseAction(s)
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
