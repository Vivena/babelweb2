package parser

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strconv"
)

var (
	errEOL                = errors.New("EOL")
	ErrUnterminatedString = errors.New("Unterminated String")
)

type Parser struct {
	scanner *bufio.Scanner

	Router  string
	Host    string
	Version string

	keywords map[string]keywordParser
}

type Transition struct {
	Router string `json:"router"`
	Host   string `json:"name"`

	Action string                 `json:"action"`
	Table  string                 `json:"table"`
	Field  string                 `json:"id"`
	Data   map[string]interface{} `json:"data"`
}

type keywordParser func(*bufio.Scanner) (interface{}, error)

func NewParser(r io.Reader) *Parser {
	scanner := bufio.NewScanner(r)
	scanner.Split(split)

	return &Parser{
		scanner: scanner,

		Router:  "",
		Host:    "unknown",
		Version: "unknown",

		keywords: map[string]keywordParser{
			"BABEL":     parseString,
			"version":   parseString,
			"host":      parseString,
			"my-id":     parseString,
			"up":        parseBool,
			"ipv4":      parseIP,
			"ipv6":      parseIP,
			"address":   parseIP,
			"if":        parseString,
			"reach":     getUintParser(16, 16),
			"ureach":    getUintParser(16, 16),
			"cost":      getUintParser(10, 32),
			"rxcost":    getUintParser(10, 32),
			"txcost":    getUintParser(10, 32),
			"rtt":       parseString,
			"rttcost":   parseString,
			"prefix":    parsePrefix,
			"from":      parsePrefix,
			"installed": parseBool,
			"id":        parseString,
			"metric":    getUintParser(10, 32),
			"refmetric": getUintParser(10, 32),
			"via":       parseIP,
		},
	}
}

func (p *Parser) Init() error {
	for {
		keyword, err := nextWord(p.scanner)
		if err != nil && err != errEOL {
			return err
		}

		parseData, exists := p.keywords[keyword]
		if !exists {
			for err != errEOL {
				_, err := nextWord(p.scanner)
				if err != nil && err != errEOL {
					return err
				}
			}
			continue
		}

		data, err := parseData(p.scanner)
		if err != nil && err != errEOL && (err != io.EOF || p.Router == "") {
			return err
		}

		switch keyword {
		case "BABEL":
			if data != nil && data.(string) == "0.0" {
				return errors.New("BABEL 0.0: Unsupported version")
			}
		case "host":
			if data != nil {
				p.Host = data.(string)
			}
		case "version":
			if data != nil {
				p.Version = data.(string)
			}
		case "my-id":
			if data != nil {
				p.Router = data.(string)
				return nil
			}
		}
	}
}

func (p *Parser) Parse() (*Transition, error) {
	action, err := nextWord(p.scanner)
	if err == errEOL {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	table, err := nextWord(p.scanner)
	if err == errEOL {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	field, err := nextWord(p.scanner)
	if err == errEOL {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})

	keyword, err := nextWord(p.scanner)
	for err != errEOL && err != io.EOF {
		if err != nil {
			return nil, err
		}

		dataParser, exists := p.keywords[keyword]
		if !exists {
			dataParser = parseString
		}

		data[keyword], err = dataParser(p.scanner)
		if err != nil {
			return nil, err
		}

		keyword, err = nextWord(p.scanner)
	}

	return &Transition{
		Router: p.Router,
		Host:   p.Host,
		Action: action,
		Table:  table,
		Field:  field,
		Data:   data,
	}, nil
}

func nextWord(s *bufio.Scanner) (string, error) {
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

func parseString(s *bufio.Scanner) (interface{}, error) {
	return nextWord(s)
}

func parseBool(s *bufio.Scanner) (interface{}, error) {
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

func getIntParser(base int, bitSize int) keywordParser {
	return func(s *bufio.Scanner) (interface{}, error) {
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

func getUintParser(base int, bitSize int) keywordParser {
	return func(s *bufio.Scanner) (interface{}, error) {
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

func parseIP(s *bufio.Scanner) (interface{}, error) {
	w, err := nextWord(s)
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(w)
	if ip == nil {
		return nil, errors.New("Syntax Error: invalid IP address: " + w)
	}

	return w, nil
}

func parsePrefix(s *bufio.Scanner) (interface{}, error) {
	w, err := nextWord(s)
	if err != nil {
		return nil, err
	}

	_, _, err = net.ParseCIDR(w)
	if err != nil {
		return nil, errors.New("Syntax Error: " + err.Error())
	}

	return w, nil
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
	for i < len(data) && data[i] != ' ' && data[i] != '\r' && data[i] != '\n' {
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
