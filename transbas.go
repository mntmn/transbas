package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

var inbuf []byte
var inbufPos = 0
var curLineNumber = 0

const (
	kwCmd  = iota
	kwFunc = iota
	kwOp   = iota
)

const (
	tInt    = iota
	tString = iota
	tReal   = iota
)

type keyword struct {
	name    string
	kwtype  int
	handler func() (bool, string)
}

type variable struct {
	name    string
	vartype int
}

func nopHandler() (bool, string) {
	return true, ""
}

func syntaxError() {
	fmt.Println("syntax error in line", curLineNumber)
}

func printHandler() (bool, string) {
	out := ""
	nl := "  printf(\"\\n\");\n"
	for {
		if ok := peekSep(); ok {
			return true, out + nl
		} else if ok, exprOut, et := readExpression(); ok {
			if et == tInt {
				out += "printf(\"%d\"," + exprOut + ");\n"
			} else if et == tString {
				out += "printf(\"%s\"," + exprOut + ");\n"
			} else if et == tReal {
				out += "printf(\"%f\"," + exprOut + ");\n"
			}
			readComma(false)
		} else {
			if c, ok := peek(); ok && c == ';' {
				// "no newline" indicator
				ffwd(1)
				nl = ""
			} else {
				break
			}
		}
	}
	return false, ""
}

func inputHandler() (bool, string) {
	out := ""
	q, _ := readString()
	out += "printf(\"" + q + "\");\n"
	readComma(true)
	v := readVar()
	variables[v.name] = *v
	out += "scanf(\"%s\",_" + v.name + ");\n"
	return true, out
}

func getHandler() (bool, string) {
	out := ""
	v := readVar()
	variables[v.name] = *v
	out += "scanf(\"%s\",_" + v.name + ");\n"
	return true, out
}

func gotoHandler() (bool, string) {
	n, ok := readNumber()
	if ok {
		return true, fmt.Sprintf("goto line%v;\n", n)
	}
	return false, ""
}

func gosubHandler() (bool, string) {
	n, ok := readNumber()
	if ok {
		return true, fmt.Sprintf("goto line%v;\n", n)
	}
	return false, ""
}

func onHandler() (bool, string) {
	out := ""
	ok, ex, _ := readExpression()
	if ok {
		out += "switch ((int)(" + ex + ")) {\n"
		readKeyword() // goto
		for {
			if n, ok := readNumber(); ok {
				out += fmt.Sprintf("goto line%v;\n", n)
				ok, _ := readComma(false)
				if !ok {
					break
				}
			} else {
				break
			}
		}
		out += "}\n"
	}
	return true, out
}

func returnHandler() (bool, string) {
	return true, "// return\n"
}

func endHandler() (bool, string) {
	return true, "return; // end\n"
}

func pokeHandler() (bool, string) {
	ok, a, _ := readExpression()
	ok2, _ := readComma(false)
	ok3, b, _ := readExpression()
	if ok && ok2 && ok3 {
		return true, fmt.Sprintf("poke(%v,%v);\n", a, b)
	}
	return false, ""
}

func sysHandler() (bool, string) {
	_, a, _ := readExpression()
	argc := 0
	for {
		if ok, _ := readComma(true); ok {
			readExpression()
			argc += 1
		} else {
			if !peekSep() {
				readExpression()
			}
			break
		}
	}

	//if ok {
	return true, fmt.Sprintf("sys(%v); // %d args\n", a, argc)
	//}
	//return false, ""
}

func waitHandler() (bool, string) {
	_, a, _ := readExpression()
	readComma(false)
	_, b, _ := readExpression()

	return true, fmt.Sprintf("// wait %s,%s\n", a, b)
}

func dataHandler() (bool, string) {
	for {
		readExpression()
		if ok, _ := readComma(false); !ok {
			break
		}
	}
	return true, fmt.Sprintf("// data\n")
}

func readHandler() (bool, string) {
	readVar()
	return true, fmt.Sprintf("// read\n")
}

func remHandler() (bool, string) {
	p := inbufPos
	for {
		if c, ok := peek(); ok {
			if c == '\n' || c == '\r' {
				ffwd(1)
				return true, ""
			} else {
				ffwd(1)
			}
		} else {
			break
		}
	}
	rewind(p)
	return false, ""
}

func tabFuncHandler() (bool, string) {
	_, ex, _ := readExpression()
	return true, "tab(" + ex + ")"
}
func freFuncHandler() (bool, string) {
	_, ex, _ := readExpression()
	return true, "abs(" + ex + ")"
}
func absFuncHandler() (bool, string) {
	_, ex, _ := readExpression()
	return true, "abs(" + ex + ")"
}
func intFuncHandler() (bool, string) {
	_, ex, _ := readExpression()
	return true, "(int)(" + ex + ")"
}
func valFuncHandler() (bool, string) {
	_, ex, _ := readExpression()
	return true, "atoi(" + ex + ")"
}
func rndFuncHandler() (bool, string) {
	_, ex, _ := readExpression()
	return true, "rand(" + ex + ")"
}
func sinFuncHandler() (bool, string) {
	_, ex, _ := readExpression()
	return true, "sin(" + ex + ")"
}
func cosFuncHandler() (bool, string) {
	_, ex, _ := readExpression()
	return true, "cos(" + ex + ")"
}
func tanFuncHandler() (bool, string) {
	_, ex, _ := readExpression()
	return true, "tan(" + ex + ")"
}
func chrFuncHandler() (bool, string) {
	_, ex, _ := readExpression()
	return true, "chr(" + ex + ")"
}
func ascFuncHandler() (bool, string) {
	_, ex, _ := readExpression()
	return true, ex + "[0]"
}
func timeFuncHandler() (bool, string) {
	return true, "jiffies()"
}
func lenFuncHandler() (bool, string) {
	_, ex, _ := readExpression()
	return true, "strlen(" + ex + ")"
}
func sgnFuncHandler() (bool, string) {
	_, ex, _ := readExpression()
	return true, "sgn(" + ex + ")"
}
func sqrFuncHandler() (bool, string) {
	_, ex, _ := readExpression()
	return true, "sqrt(" + ex + ")"
}
func peekFuncHandler() (bool, string) {
	_, ex, _ := readExpression()
	return true, "peek(" + ex + ")"
}

var keywords = map[string]keyword{}
var variables = map[string]variable{}

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
				if c, ok := peek(); ok && c == ';' {
					ffwd(1)
				}
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

func readComma(semiOk bool) (bool, bool) {
	p := inbufPos
	for {
		if c, ok := peek(); ok {
			if c == ',' || (c == ';' && semiOk) {
				ffwd(1)
				return true, c == ';'
			} else if c == ' ' {
				ffwd(1)
			} else {
				break
			}
		} else {
			break
		}
	}
	rewind(p)
	return false, false
}

func readOpenParen() (string, bool) {
	p := inbufPos
	for {
		if c, ok := peek(); ok {
			if c == '(' {
				ffwd(1)
				return string(c), true
			} else {
				break
			}
		} else {
			break
		}
	}
	rewind(p)
	return "", false
}

func readCloseParen() (string, bool) {
	p := inbufPos
	for {
		if c, ok := peek(); ok {
			if c == ')' {
				ffwd(1)
				return string(c), true
			} else {
				break
			}
		} else {
			break
		}
	}
	rewind(p)
	return "", false
}

func readNumber() (float64, bool) {
	tok := ""
	spaceok := true
	p := inbufPos
	for {
		if c, ok := peek(); ok {
			if (c >= '0' && c <= '9') || (spaceok && c == '-') || c == '.' {
				spaceok = false
				ffwd(1)
				tok = tok + string(c)
			} else if spaceok && (c == ' ' || c == '\n' || c == '\r') {
				ffwd(1)
				// skip whitespace
			} else {
				if len(tok) > 0 {
					n, _ := strconv.ParseFloat(tok, 64)
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
			if c >= 'a' && c <= 'z' || (!spaceok && c >= '0' && c <= '9') {
				spaceok = false
				ffwd(1)
				tok = tok + string(c)
			} else if spaceok && (c == ' ') {
				ffwd(1)
			} else if c == '$' && len(tok) > 0 {
				tok = tok + "Str"
				ffwd(1)

				// FIXME DRY
				if c, ok := peek(); ok && c == '(' {
					tok = "arr" + tok
					// array
					ffwd(1)
					if ok, dim, _ := readExpression(); ok {
						if ok, _ := readComma(false); ok {
							if ok, dim2, _ := readExpression(); ok {
								ffwd(1)
								tok += fmt.Sprintf("[(int)(%s)][(int)(%s)]", dim, dim2)
							}
						} else {
							ffwd(1)
							tok += fmt.Sprintf("[(int)(%s)]", dim)
						}
					}
				}

				return &variable{
					name:    tok,
					vartype: tString,
				}
			} else if c == '%' && len(tok) > 0 {
				tok = tok + "Int"
				ffwd(1)

				if c, ok := peek(); ok && c == '(' {
					tok = "arr" + tok
					// array
					ffwd(1)
					if ok, dim, _ := readExpression(); ok {
						if ok, _ := readComma(false); ok {
							if ok, dim2, _ := readExpression(); ok {
								ffwd(1)
								tok += fmt.Sprintf("[(int)(%s)][(int)(%s)]", dim, dim2)
							}
						} else {
							ffwd(1)
							tok += fmt.Sprintf("[(int)(%s)]", dim)
						}
					}
				}

				return &variable{
					name:    tok,
					vartype: tInt,
				}
			} else if len(tok) > 0 {
				// FIXME DRY
				if c, ok := peek(); ok && c == '(' {
					tok = "arr" + tok
					// array
					ffwd(1)
					if ok, dim, _ := readExpression(); ok {
						if ok, _ := readComma(false); ok {
							if ok, dim2, _ := readExpression(); ok {
								ffwd(1)
								tok += fmt.Sprintf("[(int)(%s)][(int)(%s)]", dim, dim2)
							}
						} else {
							ffwd(1)
							tok += fmt.Sprintf("[(int)(%s)]", dim)
						}
					}
				}

				return &variable{
					name:    tok,
					vartype: tReal,
				}
			} else {
				break
			}
		}
	}
	rewind(p)
	return nil
}

func dimHandler() (bool, string) {
	out := ""
	for {
		v := readVar()
		//out += fmt.Sprintf("// dim %v\n", v)

		variables[v.name] = *v
		out += "// dim " + v.name + "\n"

		if ok, _ := readComma(false); !ok {
			break
		}
	}
	return true, out
}

func forHandler() (bool, string) {
	out := ""
	v := readVar()
	step := "1"

	readOp()
	_, x1, _ := readExpression()
	readKeyword()
	_, x2, _ := readExpression()

	if kw := readKeyword(); kw != nil {
		// step
		_, x3, _ := readExpression()
		step = x3
	}

	// FIXME type inference
	v.vartype = tInt
	variables[v.name] = *v

	out += fmt.Sprintf("for (_%v=%v;_%v<=%v;_%v+=%v) {\n", v.name, x1, v.name, x2, v.name, step)

	return true, out
}

func ifHandler() (bool, string) {
	out := ""
	x2 := ""
	_, x1, _ := readExpression()
	readKeyword()
	if num, ok := readNumber(); ok {
		// goto shortcut
		x2 = fmt.Sprintf("goto line%v;\n", num)
	} else {
		_, x2 = readStatement()
	}

	out += fmt.Sprintf("if (%v) {\n    %v  }\n", x1, x2)
	return true, out
}

func nextHandler() (bool, string) {
	readVar()
	return true, "while(0){}; }\n"
}

func readOp() (string, bool) {
	p := inbufPos
	for {
		if c, ok := peek(); ok {
			if c == '<' && string(inbuf[inbufPos:inbufPos+2]) == "<>" {
				ffwd(2)
				return " != ", true
			} else if c == '<' && string(inbuf[inbufPos:inbufPos+2]) == "<=" {
				ffwd(2)
				return " <= ", true
			} else if c == '>' && string(inbuf[inbufPos:inbufPos+2]) == "=>" {
				ffwd(2)
				return " => ", true
			} else if c == '+' || c == '-' || c == '*' || c == '/' || c == '<' || c == '>' {
				ffwd(1)
				return string(c), true
			} else if c == '^' {
				ffwd(1)
				return "*", true
			} else if c == '=' {
				ffwd(1)
				return "==", true
			} else if c == 'a' && string(inbuf[inbufPos:inbufPos+3]) == "and" {
				ffwd(3)
				return " & ", true
			} else if c == 'n' && string(inbuf[inbufPos:inbufPos+3]) == "not" {
				ffwd(3)
				return " !", true
			} else if c == 'o' && string(inbuf[inbufPos:inbufPos+2]) == "or" {
				ffwd(2)
				return " | ", true
			} else if c == ' ' {
				ffwd(1)
			} else {
				break
			}
		}
	}
	rewind(p)
	return "", false
}

func readExpression() (bool, string, int) {
	state := 2
	out := ""
	exprType := tInt
	paren := 0

	for {
		if state == 0 {
			parenopened := false

			if kw := readKeyword(); kw != nil && kw.kwtype == kwFunc {
				// has to be a function
				if kw.name == "ti" || kw.name == "time" {
					// special functions without arguments
					_, ex := kw.handler()
					out += ex
				} else {
					if c, ok := peek(); ok && c == '(' {
						ffwd(1)
						//fmt.Println("  funccall: ", kw)
						// read args
						_, ex := kw.handler()
						out += ex
						//_, ex, _ := readExpression()

						if c, ok := peek(); !ok || c != ')' {
							break
						}
						ffwd(1)

						if kw.name == "chr$" && !peekSep() {
							// can be followed by a string
							if ok, _, _ := readExpression(); ok {
								//fmt.Println("--> ", ex, curLineNumber)
							}
						}
					}
				}

				// TODO et from func def
			} else if num, ok := readNumber(); ok {
				if num == float64(int(num)) {
					exprType = tInt
					out += fmt.Sprintf("%v", int(num))
				} else {
					exprType = tReal
					out += fmt.Sprintf("%v", num)
				}
			} else if str, ok := readString(); ok {
				out += fmt.Sprintf("\"%v\"", str)
				exprType = tString
			} else if v := readVar(); v != nil {
				out += fmt.Sprintf("_%v", v.name)
				exprType = v.vartype
			} else if str, ok := readOpenParen(); ok {
				out += str
				paren += 1
				parenopened = true
			} else {
				break
			}

			if paren > 0 {
				if str, ok := readCloseParen(); ok {
					out += str
					paren -= 1
				}
			}

			if !parenopened {
				state = 1
			}
		} else {
			if op, ok := readOp(); ok {
				out += op
				state = 0
			} else {
				if state == 2 {
					state = 0
				} else {
					return true, out, exprType
				}
			}
		}
	}

	return false, "", 0
}

func readAssignment() (bool, string) {
	p := inbufPos
	out := ""
	if v := readVar(); v != nil {
		// don't declare arrays (declared by dim)
		if !strings.ContainsRune(v.name, '[') {
			variables[v.name] = *v
		}

		if c, ok := peek(); ok {
			if c == '=' {
				ffwd(1)
			} else {
				rewind(p)
				return false, ""
			}
		} else {
			return false, ""
		}

		ok, ex, _ := readExpression()
		if !ok {
			rewind(p)
		} else {
			if v.vartype == tString {
				out += "strcpy(_" + v.name + "," + ex + ");"
			} else {
				out += fmt.Sprintf("_%v=%s;\n", v.name, ex)
			}
			return true, out
		}
	}
	return false, ""
}

func readKeyword() *keyword {
	tok := ""
	spaceok := true
	p := inbufPos
	for {
		if c, ok := peek(); ok {
			if (c >= 'a' && c <= 'z') || c == '$' {
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
					//fmt.Println("unknown keyword:", tok)
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

func readStatement() (bool, string) {
	out := ""
	kw := readKeyword()
	if kw != nil {
		if ok, str := kw.handler(); !ok {
			syntaxError()
			return false, ""
		} else {
			return true, out + str
		}
	} else {
		ok, str := readAssignment()
		if !ok {
			return false, ""
		}
		return true, out + str
	}
}

func readLine() (bool, string) {
	out := ""
	lineNumber := curLineNumber
	ok := false
	if sep := readSep(false); sep {
		ok = true
	} else {
		lineNumberF := 0.0
		lineNumberF, ok = readNumber()
		lineNumber = int(lineNumberF)
	}

	if ok {
		if curLineNumber != lineNumber {
			out += fmt.Sprintf("\nline%v:\n  ", lineNumber)
		}
		curLineNumber = lineNumber
		ok, out2 := readStatement()
		return ok, out + out2
	}
	return false, ""
}

func main() {

	keywords = map[string]keyword{
		"let":    keyword{"let", kwCmd, nopHandler},
		"goto":   keyword{"goto", kwCmd, gotoHandler},
		"gosub":  keyword{"gosub", kwCmd, gosubHandler},
		"on":     keyword{"on", kwCmd, onHandler},
		"return": keyword{"return", kwCmd, returnHandler},
		"end":    keyword{"end", kwCmd, returnHandler},
		"print":  keyword{"print", kwCmd, printHandler},
		"input":  keyword{"input", kwCmd, inputHandler},
		"get":    keyword{"get", kwCmd, getHandler},
		"poke":   keyword{"poke", kwCmd, pokeHandler},
		"sys":    keyword{"sys", kwCmd, sysHandler},
		"dim":    keyword{"dim", kwCmd, dimHandler},
		"for":    keyword{"for", kwCmd, forHandler},
		"if":     keyword{"if", kwCmd, ifHandler},
		"next":   keyword{"next", kwCmd, nextHandler},

		"rem":  keyword{"rem", kwCmd, remHandler},
		"new":  keyword{"new", kwCmd, nopHandler},
		"clr":  keyword{"clr", kwCmd, nopHandler},
		"list": keyword{"list", kwCmd, nopHandler},
		"to":   keyword{"to", kwCmd, nopHandler},
		"then": keyword{"then", kwCmd, nopHandler},
		"step": keyword{"step", kwCmd, nopHandler},
		"wait": keyword{"wait", kwCmd, waitHandler},
		"data": keyword{"data", kwCmd, dataHandler},
		"read": keyword{"read", kwCmd, readHandler},

		"abs":  keyword{"abs", kwFunc, absFuncHandler},
		"int":  keyword{"int", kwFunc, intFuncHandler},
		"val":  keyword{"val", kwFunc, valFuncHandler},
		"tab":  keyword{"tab", kwFunc, tabFuncHandler},
		"spc":  keyword{"spc", kwFunc, tabFuncHandler},
		"fre":  keyword{"fre", kwFunc, freFuncHandler},
		"rnd":  keyword{"rnd", kwFunc, rndFuncHandler},
		"sin":  keyword{"sin", kwFunc, sinFuncHandler},
		"cos":  keyword{"cos", kwFunc, cosFuncHandler},
		"tan":  keyword{"tan", kwFunc, tanFuncHandler},
		"ti":   keyword{"ti", kwFunc, timeFuncHandler},
		"time": keyword{"time", kwFunc, timeFuncHandler},
		"atn":  keyword{"atn", kwFunc, tanFuncHandler},
		"len":  keyword{"len", kwFunc, lenFuncHandler},
		"peek": keyword{"peek", kwFunc, peekFuncHandler},

		"chr$": keyword{"chr$", kwFunc, chrFuncHandler},
		"asc":  keyword{"asc", kwFunc, ascFuncHandler},
		"sgn":  keyword{"sgn", kwFunc, sgnFuncHandler},
		"sqr":  keyword{"sqr", kwFunc, sqrFuncHandler},
	}

	ib, err := ioutil.ReadFile("test.txt")
	if err != nil {
		fmt.Println("couldn't open test.txt")
		return
	}
	inbuf = ib

	var lines []string

	fmt.Println("#include <stdio.h>")
	fmt.Println("#include <string.h>")
	fmt.Println("#include <math.h>")
	fmt.Println("void poke(int a, char b) {}")
	fmt.Println("void sys(int a) {}")
	fmt.Println("char* tab(int a) {return \" \";}")
	fmt.Println("char chr(int a) {return a;}")

	fmt.Println("int rand(int a) {return 1;}")
	fmt.Println("int abs(int a) {return 1;}")
	fmt.Println("int tan(int a) {return 1;}")
	fmt.Println("int atn(int a) {return 1;}")
	fmt.Println("int sqrt(int a) {return 1;}")
	fmt.Println("int jiffies() {return 1;}")

	for {
		if ok, str := readLine(); ok {
			lines = append(lines, str)
		} else {
			break
		}
	}

	for _, v := range variables {
		if v.vartype == tString {
			fmt.Println("char _" + strings.Replace(v.name, "(int)", "", 2) + "[24];")
		} else if v.vartype == tInt {
			fmt.Println("int _" + strings.Replace(v.name, "(int)", "", 2) + ";")
		} else if v.vartype == tReal {
			fmt.Println("int _" + strings.Replace(v.name, "(int)", "", 2) + ";")
		}
	}

	fmt.Println("void main() {")

	for _, line := range lines {
		fmt.Print("  " + line)
	}
	fmt.Println("}")
}
