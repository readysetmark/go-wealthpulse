package parse

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// identifies the type of lex items
type itemType int

// represents a token returned from the scanner
type item struct {
	typ   itemType
	value string
}

const (
	// error occurred; value is text of error
	itemError itemType = iota
	itemPriceSentinel
	itemYear
	itemMonth
	itemDayOfMonth
	itemQuotedUnit
	itemUnit
	itemQuantity
	itemEOF
	itemTODO
)

const (
	eof           = 0
	priceSentinel = "P"
)

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.value
	}
	return fmt.Sprintf("%q", i.value)
}

// holds the state of the scanner
type lexer struct {
	name  string    // used for error reports
	input string    // input being scanned
	state stateFn   // the current stateFn
	start int       // start position of this item
	pos   int       // current position in the input
	width int       // width of the last rune read
	items chan item // channel of scanned items
}

func lex(name, input string, startState stateFn) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		state: startState,
		items: make(chan item, 2), // channel with buffer size of 2
	}
	return l
}

// returns the next item from the input
func (l *lexer) nextItem() item {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			l.state = l.state(l)
		}
	}
	//panic("not reached")
}

// passes an item back to the client
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

// returns the next rune in the input
func (l *lexer) next() (rune rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	rune, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return rune
}

// skips over the pending input to this point
func (l *lexer) ignore() {
	l.start = l.pos
}

// steps back one rune. can be called only once per call of next
func (l *lexer) backup() {
	l.pos -= l.width
}

// returns but does not consume the next input rune
func (l *lexer) peek() rune {
	rune := l.next()
	l.backup()
	return rune
}

// consumes the next rune if it's from the valid set
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// consumes a run of runes from the valid set
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// returns an error token and terminates the scan by passing back a nil stateFn
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, args...),
	}
	return nil
}

// represents the state of the scanner as a function that returns next state
type stateFn func(*lexer) stateFn

//
// My helpers
//

// ignore the next rune (whatever it is) then go to the next stateFn
func ignoreNext(nextState stateFn) stateFn {
	return func(l *lexer) stateFn {
		l.next()
		l.ignore()
		return nextState
	}
}

//
// Wealthpulse-specific stuff begins here
//

// THIS IS JUST A SAMPLE -- lex to the end!
func lexAll(l *lexer) stateFn {
	for {
		if l.next() == eof {
			break
		}
	}
	// correctly reached EOF
	if l.pos > l.start {
		l.emit(itemTODO)
	}
	l.emit(itemEOF)
	return nil
}

func lexPrice(l *lexer) stateFn {
	return lexPriceSentinel
}

func lexPriceSentinel(l *lexer) stateFn {
	l.accept("P")
	l.emit(itemPriceSentinel)
	return ignoreNext(lexYear)
}

func lexYear(l *lexer) stateFn {
	l.pos += 4
	l.emit(itemYear)
	return ignoreNext(lexMonth)
}

func lexMonth(l *lexer) stateFn {
	l.pos += 2
	l.emit(itemMonth)
	return ignoreNext(lexDayOfMonth)
}

func lexDayOfMonth(l *lexer) stateFn {
	l.pos += 2
	l.emit(itemDayOfMonth)
	return ignoreNext(lexQuotedUnit)
}

func lexQuotedUnit(l *lexer) stateFn {
	l.accept("\"")
	l.ignore()
	for strings.IndexRune("\"\r\n", l.next()) == -1 {
	}
	l.backup()
	l.emit(itemQuotedUnit)
	l.accept("\"")
	l.ignore()
	return ignoreNext(lexAmount)
}

func lexAmount(l *lexer) stateFn {
	return lexUnit
}

func lexUnit(l *lexer) stateFn {
	for strings.IndexRune("-0123456789; \"\t\r\n", l.next()) == -1 {
	}
	l.backup()
	l.emit(itemUnit)
	return lexQuantity
}

func lexQuantity(l *lexer) stateFn {
	l.accept("-")
	l.acceptRun("0123456789")
	l.accept(".")
	l.acceptRun("0123456789")
	l.emit(itemQuantity)
	return lexAfterPrice
}

func lexAfterPrice(l *lexer) stateFn {
	l.acceptRun("\r\n")
	l.ignore()
	next := l.peek()
	if next == eof {
		l.emit(itemEOF)
		return nil
	}
	return lexPrice
}

func lexPriceDB(l *lexer) stateFn {
	next := l.peek()
	if next == eof {
		return nil
	}
	return lexPrice
}
