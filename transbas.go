package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
)

var inbuf []byte
var inbufPos = 0
var curLineNumber = 0

type keyword struct {
	name    string
	handler func() bool
}

type variable struct {
	name    string
	vartype string
}

func nopHandler() bool {
	return true
}

func syntaxError() {
	fmt.Println("syntax error in line", curLineNumber)
}

func printHandler() bool {
	for {
		if ok := peekSep(); ok {
			return true
		} else if str, ok := readString(); ok {
			fmt.Println("print string:", str)
		} else if v := readVar(); v != nil {
			fmt.Println("print var:", v)
		} else {
			break
		}
	}
	return false
}

func gotoHandler() bool {
	n, ok := readNumber()
	if ok {
		fmt.Println("goto:", n)
		return true
	}
	return false
}

var keywords = map[string]keyword{
	"REM":   keyword{"rem", nopHandler},
	"LET":   keyword{"let", letHandler},
	"GOTO":  keyword{"goto", gotoHandler},
	"PRINT": keyword{"print", printHandler},
}

func ffwd(n int) {
	inbufPos += n
}

func rewind(n int) {
	inbufPos = n
}

func peek() (byte, bool) {
	if inbufPos >= len(inbuf) {
		return 0, false
	}
	return inbuf[inbufPos], true
}

func peekSep() bool {
	p := inbufPos
	sep := readSep(true)
	rewind(p)
	return sep
}

func readSep(withNewline bool) bool {
	p := inbufPos
	for {
		if c, ok := peek(); ok {
			if c == ':' {
				ffwd(1)
				return true
			} else if withNewline && (c == '\n' || c == '\r') {
				ffwd(1)
				return true
			} else if c == ' ' {
				ffwd(1)
			} else {
				break
			}
		}
	}
	rewind(p)
	return false
}

func readString() (string, bool) {
	tok := ""
	spaceok := true
	p := inbufPos
	for {
		if c, ok := peek(); ok {
			if c == '"' && spaceok {
				spaceok = false
				ffwd(1)
			} else if c == '"' && !spaceok {
				ffwd(1)
				return tok, true
			} else if spaceok && (c == ' ') {
				ffwd(1)
			} else if !spaceok {
				tok = tok + string(c)
				ffwd(1)
			} else {
				break
			}
		}
	}
	rewind(p)
	return "", false
}

func readNumber() (int, bool) {
	tok := ""
	spaceok := true
	p := inbufPos
	for {
		if c, ok := peek(); ok {
			if c >= '0' && c <= '9' {
				spaceok = false
				ffwd(1)
				tok = tok + string(c)
			} else if spaceok && (c == ' ' || c == '\n' || c == '\r') {
				ffwd(1)
				// skip whitespace
			} else {
				if len(tok) > 0 {
					n, _ := strconv.Atoi(tok)
					return n, true
				}
				break
			}
		} else {
			break
		}
	}
	rewind(p)
	return 0, false
}

func readVar() *variable {
	tok := ""
	spaceok := true
	p := inbufPos
	for {
		if c, ok := peek(); ok {
			if c >= 'A' && c <= 'Z' {
				spaceok = false
				ffwd(1)
				tok = tok + string(c)
			} else if spaceok && (c == ' ') {
				ffwd(1)
			} else if c == '$' && len(tok) > 0 {
				tok = tok + string(c)
				ffwd(1)
				return &variable{
					name:    tok,
					vartype: "string",
				}
			} else if c == '%' && len(tok) > 0 {
				tok = tok + string(c)
				ffwd(1)
				return &variable{
					name:    tok,
					vartype: "int",
				}
			} else if len(tok) > 0 {
				return &variable{
					name:    tok,
					vartype: "real",
				}
			} else {
				break
			}
		}
	}
	rewind(p)
	return nil
}

func readExpression() bool {
	return false
}

func readAssignment() bool {
	p := inbufPos
	readVar()
	if c, ok := peek(); ok {
		if c == '=' {
			return true
		} else {
			return false
		}
	}
	ok := readExpression()
	return ok
}

func readKeyword() *keyword {
	tok := ""
	spaceok := true
	p := inbufPos
	for {
		if c, ok := peek(); ok {
			if c >= 'A' && c <= 'Z' {
				spaceok = false
				ffwd(1)
				tok = tok + string(c)
				if kw, ok := keywords[tok]; ok {
					return &kw
				}
			} else if spaceok && (c == ' ' || c == '\n' || c == '\r') {
				ffwd(1)
				// skip whitespace
			} else {
				if len(tok) > 0 {
					fmt.Println("unknown keyword:", tok)
				}
				break
			}
		} else {
			break
		}
	}
	rewind(p)
	return nil
}

func readLine() bool {
	lineNumber := curLineNumber
	ok := false
	if sep := readSep(false); sep {
		ok = true
	} else {
		lineNumber, ok = readNumber()
	}

	if ok {
		fmt.Println("lineno:", lineNumber)
		curLineNumber = lineNumber
		kw := readKeyword()
		fmt.Println("keyword:", kw)
		if kw != nil {
			if !kw.handler() {
				syntaxError()
				return false
			}
		}
		return true
	}
	return false
}

func main() {
	ib, err := ioutil.ReadFile("test.txt")
	if err != nil {
		fmt.Println("couldn't open test.txt")
		return
	}
	inbuf = ib
	for readLine() {
	}
}
