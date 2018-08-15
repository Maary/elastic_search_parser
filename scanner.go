package parser

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strings"
)

//`LOOK (index1'tpe, index2'tpe): [field1, field2, field3]  CONDITION: {1.field1 GT 100, 1.field2 PF "prefix", 2.field3 SF "suffix"}  AT: {1.begin TO 1.end, 2.being TO 2.end}`.

// Scanner represents a lexical scanner.
type Scanner struct {
	r *bufio.Reader
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// Scan returns the next token and literal value.
func (s *Scanner) Scan() (tok Token, lit string) {
	// Read the next rune.
	ch := s.read()

	// If we see whitespace then consume all contiguous whitespace.
	// If we see a letter then consume as an ident or reserved word.
	// If we see a digit then consume as a number.
	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) {
		s.unread()
		return s.scanIdent()
	} else if isDigit(ch) {
		s.unread()
		return s.scanInteger()
	}

	// Otherwise read the individual character.
	switch ch {
	case eof:
		return EOF, ""
	case '(':
		return ParLeft, string(ch)
	case ')':
		return ParRight, string(ch)
	case ',':
		return COMMA, string(ch)
	case '\'':
		return OWN, string(ch)
	case ':':
		return IS, string(ch)
	case '{':
		return BParLeft, string(ch)
	case '}':
		return BParRight, string(ch)
	case '[':
		return MParLeft, string(ch)
	case ']':
		return MParRight, string(ch)
	case '.':
		return Point, string(ch)
	case '"':
		return STR, string(ch)
	case '-':
		return MIDEND, string(ch)
	case '>':
		return PointRight, string(ch)
	}

	return ILLEGAL, string(ch)
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return WS, buf.String()
}

func (s *Scanner) scanInteger() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
			// return WS, buf.String()
		} else if !isDigit(ch) && ch != '.' {
			s.unread()
			break
		} else {
			// fmt.Println("=>", string(ch))
			buf.WriteRune(ch)
		}
	}
	return IDENT, buf.String()
}

// scanIdent consumes the current rune and all contiguous ident runes.
func (s *Scanner) scanIdent() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) && ch != '_' {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// If the string matches a keyword then return that keyword.
	switch strings.ToUpper(buf.String()) {
	case "LOOK":
		return LOOK, buf.String()
	case "TOTAl":
		return TOTAL, buf.String()
	case "CONDITION":
		return CONDITION, buf.String()
	case "AT":
		return AT, buf.String()
	case "EQ":
		return EQ, buf.String()
	case "NEQ":
		return NEQ, buf.String()
	case "PF":
		return PF, buf.String()
	case "SF":
		return SF, buf.String()
	case "LT":
		return LT, buf.String()
	case "GT":
		return GT, buf.String()
	case "GTE":
		return GTE, buf.String()
	case "LTE":
		return LTE, buf.String()
	}

	// Otherwise return as a regular identifier.
	return IDENT, buf.String()
}

// read reads the next rune from the buffered reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() { _ = s.r.UnreadRune() }

// isWhitespace returns true if the rune is a space, tab, or newline.
func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' || ch == '\n' }

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool { return (ch >= '0' && ch <= '9') }

func isSymbol(ch rune) bool {
	s := string(ch)
	reg := regexp.MustCompile(`[\:\.\'\(\)\[\]\{\}]`)
	return reg.MatchString(s)
}

// eof represents a marker rune for the end of the reader.
var eof = rune(0)
