//package parser
package main

import (
	"fmt"
	//"log"
	"bufio"
	"strings"
)

/*
const (
	opAdd    = iota
	opChange = iota
)
*/

type Queue []byte

func (q *Queue)push(b byte) {
	*q = append(*q, b)
}

func (q *Queue)pop() (byte, bool) {
	if len(*q) == 0 {
		return 0, true
	}
	b := (*q)[0]
	*q = (*q)[1:len(*q)]
	return b, false
}

func (q *Queue)peek() (byte, bool) {
	if len(*q) == 0 {
		return 0, true
	}
	return (*q)[0], false
}

func (q *Queue)clean() {
	*q = nil
}

// S for superior
type SReader struct {
	reader *bufio.Reader
	queue Queue
	checkp bool
}

func (r *SReader)checkpoint() {
	r.checkp = true
}

func (r *SReader)reset() {
	r.checkp = false
}

// "go on" or "goon" ?
func (r *SReader)goOn() {
	r.checkp = false
	r.queue.clean()
}

func (r *SReader)readByte() (byte, error) {
	if r.checkp {
		b, empty := r.queue.peek()
		if empty {
			b, err := r.reader.ReadByte()
			r.queue.push(b)
			return b, err
		}
		return b, nil
		
	}
	b, empty := r.queue.pop()
	if empty {
		return r.reader.ReadByte()
	}
	return b, nil
}

func main() {
	var q Queue
	q.push(42)
	b, _ := q.peek()
	q.push(11)
	q.push(12)
	b, _ = q.pop()
	fmt.Println(b)
	b, _ = q.pop()
	fmt.Println(b)
	q.clean()
	q.push(11)
	b, _ = q.pop()
	fmt.Println(b)
	b, _ = q.pop()
	fmt.Println(b)

	fmt.Println("--------------")

	r := strings.NewReader("huge test of mine")
	var sr SReader
	sr.reader = bufio.NewReader(r);
	sr.checkp = false
	for {
		l, err := sr.readByte()
		if err != nil {
			break;
		}
		fmt.Printf("%c", l)
	}
}
/*

const sep := ' ';

func (reader *bufio.Reader)getc (byte, bool, error) {
	b, err := reader.ReadByte()
	return b, b == sep, err
}

func (reader *bufio.Reader)ungetc (error) {
	return reader.UnreadByte()
}

func parseWord(reader *bufio.Reader, word []byte) {
	var readChar func(int) (int, bool)
	readChar = func(i int) (int, bool) {

	}
}

// add, change, ...
func parseOp(reader *bufio.Reader) (op int, err error) {
	b, err, op := reader.ReadByte(), 0
	if err != nil {
		return
	}
	if b != 'a' || b != 'c' {
		err = reader.UnreadByte()
		return 0, err
	}
}
*/
