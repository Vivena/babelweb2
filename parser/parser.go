package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"strconv"
	"sync"
)

var (
	errEOL                = errors.New("EOL")
	ErrUnterminatedString = errors.New("Unterminated String")
)

type Scanner struct {
	bufio.Scanner
}

func nextWord(s *Scanner) (string, error) {
	if s.Scan() {
		word := s.Text()
		if word == "\n" {
			return "", errEOL
		}
		return word, nil
	}
	if err := s.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}

type entryParser func(*Scanner) (interface{}, error)

type entryValue struct {
	data   interface{}
	parser entryParser
}

type entry map[string](*entryValue)

func (e entry) String() string {
	b := new(bytes.Buffer)
	for id, ev := range e {
		fmt.Fprintf(b, "\t%s: %v\n", id, ev.data)
	}
	return b.String()
}

type entryError int

const (
	FieldPresence entryError = 0
	FieldAbsence  entryError = 1
)

func (e entryError) Error() string {
	if e == FieldPresence {
		return "Field Already Exists"
	} else if e == FieldAbsence {
		return "No Such Field"
	}
	return "Error of Lack of Error"
}

func NewEntry() entry {
	return make(map[string](*entryValue))
}

func (e *entry) AddField(id string, parser entryParser) error {
	_, exists := (*e)[id]
	if exists {
		return FieldPresence
	}
	(*e)[id] = new(entryValue)
	(*e)[id].data = nil
	(*e)[id].parser = parser
	return nil
}

func (e *entry) GetData(id string) (interface{}, error) {
	value, exists := (*e)[id]
	if !exists {
		return nil, FieldAbsence
	}
	return value.data, nil
}

func (e *entry) Parse(s *Scanner) error {
	for {
		w, err := nextWord(s)
		if err != nil {
			return err
		}
		value, exists := (*e)[w]
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

func NewInterfaceEntry() entry {
	i := NewEntry()
	i.AddField("up", ParseBool)
	i.AddField("ipv4", ParseIp)
	i.AddField("ipv6", ParseIp)
	return i
}

func NewNeighbourEntry() entry {
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

func NewRouteEntry() entry {
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

func NewXrouteEntry() entry {
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
func GetIntParser(base int, bitSize int) entryParser {
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
func GetUintParser(base int, bitSize int) entryParser {
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

type table struct {
	dict map[string](entry)
	sync.Mutex
}

func (t *table) String() string {
	b := new(bytes.Buffer)
	t.Lock()
	defer t.Unlock()
	for id, e := range t.dict {
		fmt.Fprintf(b, "%s: %v\n", id, e)
	}
	return b.String()
}

type BabelDesc struct {
	ID      string
	Name    string
	Version string
	ts      map[string](*table)
}

func (bd *BabelDesc) String() string {
	b := new(bytes.Buffer)
	for id, t := range bd.ts {
		fmt.Fprintf(b, "*\t%s\n", id)
		fmt.Fprintln(b, t)
	}
	return b.String()
}

func NewBabelDesc() *BabelDesc {
	ts := make(map[string](*table))
	ts["route"] = &table{dict: make(map[string](entry))}
	ts["xroute"] = &table{dict: make(map[string](entry))}
	ts["interface"] = &table{dict: make(map[string](entry))}
	ts["neighbour"] = &table{dict: make(map[string](entry))}
	return &BabelDesc{ts: ts}
}

func (t *table) Add(id string, e entry) error {
	t.Lock()
	defer t.Unlock()
	_, exists := t.dict[id]
	if exists {
		return FieldPresence
	}
	t.dict[id] = e
	return nil
}

func (t *table) Change(id string, e entry) error {
	t.Lock()
	defer t.Unlock()
	_, exists := t.dict[id]
	if !exists {
		return FieldAbsence
	}
	t.dict[id] = e
	return nil
}

func (t *table) Flush(id string) error {
	t.Lock()
	defer t.Unlock()
	_, exists := t.dict[id]
	if !exists {
		return FieldAbsence
	}
	delete(t.dict, id)
	return nil
}

type BabelUpdate struct {
	name      string
	router    string
	action    string
	table     string
	entry     string
	entryData entry
}

func (u BabelUpdate) ID() string {
	return u.router
}

type SBabelUpdate struct {
	Name      string                 `json:"name"`
	Router    string                 `json:"router"`
	Action    string                 `json:"action"`
	Table     string                 `json:"table"`
	Entry     string                 `json:"id"`
	EntryData map[string]interface{} `json:"data"`
}

func (bd *BabelDesc) Iter(f func(BabelUpdate) error) error {
	for tk, tv := range bd.ts {
		tv.Lock()
		for ek, ev := range tv.dict {
			err := f(BabelUpdate{name: bd.Name, router: bd.ID,
				action: "add", table: tk,
				entry: ek, entryData: ev})
			if err != nil {
				tv.Unlock()
				return err
			}
		}
		tv.Unlock()
	}
	return nil
}

func (upd BabelUpdate) ToSUpdate() SBabelUpdate {
	s_upd := SBabelUpdate{upd.name, upd.router, upd.action,
		upd.table, upd.entry, make(map[string]interface{})}
	for id, ev := range upd.entryData {
		switch t := ev.data.(type) {
		case *net.IPNet:
			s_upd.EntryData[id] = t.String()
		case net.IP:
			s_upd.EntryData[id] = t.String()
		default:
			s_upd.EntryData[id] = ev.data
		}
	}
	return s_upd
}

var emptyUpdate = BabelUpdate{action: "none"}

func (upd BabelUpdate) String() string {
	return fmt.Sprintf("%s: %s %s\n%s", upd.action, upd.table,
		upd.entry, upd.entryData)
}

func makeEntry(id string) (entry, error) {
	switch id {
	case "interface":
		return NewInterfaceEntry(), nil
	case "neighbour":
		return NewNeighbourEntry(), nil
	case "route":
		return NewRouteEntry(), nil
	case "xroute":
		return NewXrouteEntry(), nil
	}
	return nil, errors.New("Unknown table Id")
}

func (bd *BabelDesc) ParseAction(s *Scanner) (BabelUpdate, error) {
	w, err := nextWord(s)
	if err != nil {
		return emptyUpdate, err
	}
	if w != "add" && w != "change" && w != "flush" {
		return emptyUpdate, nil
	}
	table, err := nextWord(s)
	if err != nil {
		return emptyUpdate, err
	}
	entry, err := nextWord(s)
	if err != nil {
		return emptyUpdate, err
	}
	new_entry, err := makeEntry(table)
	if err != nil {
		return emptyUpdate, err
	}
	err = new_entry.Parse(s)
	if err != io.EOF && err != errEOL {
		return emptyUpdate, err
	}
	return BabelUpdate{name: bd.Name, router: bd.ID,
		action: w, table: table,
		entry: entry, entryData: new_entry}, err
}

func (bd *BabelDesc) Update(upd BabelUpdate) error {
	switch upd.action {
	case "add":
		return (bd.ts)[upd.table].Add(
			upd.entry, upd.entryData)
	case "change":
		return (bd.ts)[upd.table].Change(
			upd.entry, upd.entryData)
	case "flush":
		return (bd.ts)[upd.table].Flush(upd.entry)
	}
	return nil
}

func (bd *BabelDesc) CheckUpdate(upd BabelUpdate) bool {
	if upd.action != "change" {
		return true
	}
	table := (bd.ts)[upd.table]
	table.Lock()
	defer table.Unlock()
	for key, value := range table.dict[upd.entry] {
		if !(reflect.DeepEqual((*upd.entryData[key]).data, (*value).data)) {
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

func (bd *BabelDesc) Fill(s *Scanner) error {
	e := NewEntry()
	e.AddField("BABEL", ParseString)
	e.AddField("version", ParseString)
	e.AddField("host", ParseString)
	e.AddField("my-id", ParseString)
	for e["my-id"].data == nil {
		err := e.Parse(s)
		if err != nil && err != errEOL &&
			(err != io.EOF || e["my-id"].data == nil) {
			return err
		}
	}
	if e["BABEL"].data != nil && e["BABEL"].data.(string) == "0.0" {
		return errors.New("BABEL 0.0: Unsupported version")
	}
	bd.Version = e["version"].data.(string)
	bd.ID = e["my-id"].data.(string)
	if e["host"].data == nil {
		bd.Name = "unknown"
	} else {
		bd.Name = e["host"].data.(string)
	}
	return nil
}

func (bd *BabelDesc) Listen(s *Scanner, updChan chan SBabelUpdate) error {
	defer bd.Clean(updChan)
	for {
		upd, err := bd.ParseAction(s)
		if err != nil && err != io.EOF && err != errEOL {
			return err
		}
		if err == io.EOF {
			break
		}
		if upd.action != emptyUpdate.action {
			if !(bd.CheckUpdate(upd)) {
				continue
			}
			err = bd.Update(upd)
			if err != nil {
				return err
			}
			updChan <- upd.ToSUpdate()
		}
	}
	return nil
}

func (bd *BabelDesc) Clean(updChan chan SBabelUpdate) error {
	return bd.Iter(func(u BabelUpdate) error {
		u.action = "flush"
		updChan <- u.ToSUpdate()
		return nil
	})
}
