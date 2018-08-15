package parser

import (
	"fmt"
	"io"
	"strconv"
)

// Token represents a lexical token.
type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	WS

	// Literals
	IDENT // main

	// Misc characters
	OWN        // '
	COMMA      // ,
	ParLeft    // (
	ParRight   // )
	IS         // :
	BParLeft   // {
	BParRight  // }
	MParLeft   // [
	MParRight  // ]
	Point      //.
	STR        //"
	MIDEND     //-
	PointRight //>

	// Keywords
	LOOK
	TOTAL
	CONDITION
	AT
	EQ
	NEQ
	PF
	SF
	GT
	GTE
	LT
	LTE
)

// SelectStatement represents a SQL SELECT statement.
type SelectStatement struct {
	IndexToTypeSet  []map[string]string
	IndexToFieldSet []map[string]*Operation
	TimeBegin       string
	TimeEnd         string
}

type Operation struct {
	FieldName string
	Opt       string
	Value     interface{}
	ValueType string
}

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// Parse parses a SQL SELECT statement.
func (p *Parser) Parse() (*SelectStatement, error) {
	indesToType := make(map[string]string)
	stmt := &SelectStatement{}

	// First token should be a "SELECT" keyword.
	tok, lit := p.scanIgnoreWhitespace()
	if tok != LOOK && tok != TOTAL {
		p.unscan()
		return nil, fmt.Errorf("found %q, expected LOOK or TOTAL", lit)
	}
	switch tok {
	case LOOK:
		{

			tok, lit = p.scanIgnoreWhitespace()
			if tok != ParLeft {
				p.unscan()
				return nil, fmt.Errorf("found %q, expected index_name", lit)
			}

			for {
				var indexName = ""
				var tpe string
				tok, lit := p.scanIgnoreWhitespace()
				// fmt.Println("=> index begin ", tok, lit)
				if tok != IDENT && tok != ParRight && tok != MIDEND {
					p.unscan()
					return nil, fmt.Errorf("found %q expected Index name or )", lit)
				}
				indexName += lit
				tok, lit = p.scanIgnoreWhitespace()
				// fmt.Println("=> index e", tok, lit)
				if tok != OWN && tok != MIDEND {
					p.unscan()
					return nil, fmt.Errorf("found %q expected - or '", lit)
				}
				if tok == MIDEND {
					indexName += "-"
				}
				// fmt.Println("=> mid ", tok, lit)
				switch tok {
				case MIDEND:
					{
						for {
							tok, lit = p.scanIgnoreWhitespace()
							// fmt.Println("=> mid and integer ", tok, lit)
							if tok != IDENT && tok != MIDEND {
								p.unscan()
								// return nil, fmt.Errorf("found %q expect indexname", lit)
								break
							}
							if tok == MIDEND {
								indexName += "-"
							}
							if tok == IDENT {
								indexName += lit
							}
							indexName += lit
							// fmt.Println("-->", indexName)
						}
						tok, lit = p.scanIgnoreWhitespace()
						if tok != OWN {
							p.unscan()
							return nil, fmt.Errorf("found %q expect end symbol '", lit)
						}
						goto TYPE

					}
				case OWN:
					goto TYPE
				}
			TYPE:
				{
					tok, lit = p.scanIgnoreWhitespace()
					if tok != IDENT {
						p.unscan()
						return nil, fmt.Errorf("found %q expecter type name", lit)
					}
					tpe = lit
					// fmt.Println("--> typename", tpe)
					indesToType[indexName] = tpe
					tok, _ = p.scanIgnoreWhitespace()
					if tok != COMMA && tok != ParRight {
						p.unscan()
						return nil, fmt.Errorf("found %q expecter , or )", lit)
					}
					if tok == ParRight {
						stmt.IndexToTypeSet = append(stmt.IndexToTypeSet, indesToType)
						break
					}
				}

			}

			//***********************
			//CONDITION//
			//**********************
			//`LOOK (index1'tpe, index2'tpe): CONDITION [index1.field1 GT 100, index1.field2 PF "prefix", 2.field3 SF "suffix"}  AT {1.begin TO 1.end, 2.being TO 2.end]`
			IndexToFieldMap := make(map[string]*Operation)

			toc, lit := p.scanIgnoreWhitespace()
			// fmt.Println("=> after", toc, lit)
			if toc != IS {
				p.unscan()
				return nil, fmt.Errorf("found %q expecter : ", lit)
			}

			toc, lit = p.scanIgnoreWhitespace()
			// fmt.Println("=> key word ", toc, lit)
			if toc != CONDITION {
				p.unscan()
				return nil, fmt.Errorf("found %q expecter CONDITION", lit)
			}

			toc, lit = p.scanIgnoreWhitespace()
			// fmt.Println("=> con [ ", toc, lit)
			if toc != MParLeft {
				p.unscan()
				return nil, fmt.Errorf("found %q expecter [", lit)
			}

			for {
				var conditionIndexName string

				O := new(Operation)
				toc, lit = p.scanIgnoreWhitespace()
				// fmt.Println("=> con index", toc, lit)
				conditionIndexName += lit
				// fmt.Println("con index name: ", conditionIndexName)
				if toc != IDENT && toc != MParRight {
					p.unscan()
					return nil, fmt.Errorf("found %q expecter ]", lit)
				}

				toc, lit = p.scanIgnoreWhitespace()
				// fmt.Println("=> field sep", toc, lit)
				if toc != MIDEND && toc != OWN {
					p.unscan()
					return nil, fmt.Errorf("found %q expect type declare - or Point", lit)
				}
				switch toc {
				case MIDEND:
					{
						conditionIndexName += "-"
						for {
							toc, lit = p.scanIgnoreWhitespace()
							// fmt.Println("=> -field ", toc, lit)
							if toc != MIDEND && toc != IDENT {
								p.unscan()
								// return nil, fmt.Errorf("found %q expect field name", lit)
								break
							}
							if toc == MIDEND {
								conditionIndexName += "-"
							}
							if toc == IDENT {
								conditionIndexName += lit
							}
							// fmt.Println("=> filed re ", toc, lit)
						}
						toc, lit = p.scanIgnoreWhitespace()
						if toc != OWN {
							p.unscan()
							return nil, fmt.Errorf("found %q expect end symbol", lit)
						}
						goto FIELD
					}
				case OWN:
					goto FIELD
				}

				//*
			FIELD:
				{
					// fmt.Println("--> cond index: ", conditionIndexName)
					//*
					/*
						toc, lit = p.scan()
						if toc != Point {
							p.unscan()
							return nil, fmt.Errorf("found %q expecter", lit)
						}
					*/
					toc, lit = p.scan()
					if toc != IDENT {
						p.unscan()
						return nil, fmt.Errorf("found %q expecter", lit)
					}
					//*
					O.FieldName = lit
					//*
					toc, lit = p.scanIgnoreWhitespace()
					if toc != EQ && toc != NEQ && toc != PF && toc != SF && toc != LT && toc != GT && toc != GTE && toc != LTE {
						p.unscan()
						return nil, fmt.Errorf("found %q expecter Opt: GT or LT or PF ...", lit)
					}
					I2O := make(map[string]*Operation)
					I2OSet := make([]map[string]*Operation, 0)
					// fmt.Println("=> case:", toc, lit)
					switch toc {
					case GT:
						{
							//*
							O.Opt = "GT"
							//*
							gtTocNext, gtLitNext := p.scanIgnoreWhitespace()
							// fmt.Println("=> gt int value", gtTocNext, gtLitNext)
							if gtTocNext != IDENT {
								p.unscan()
								return nil, fmt.Errorf("found %q expecter int value", gtLitNext)
							}
							gtLitNextFloat, err := strconv.ParseFloat(gtLitNext, 64)
							if err != nil {
								panic(err)
							}
							//*
							O.Value = gtLitNextFloat
							O.ValueType = "Float64"
							//*
							I2O[conditionIndexName] = O
							I2OSet = append(I2OSet, I2O)
						}
					case GTE:
						{
							//*
							O.Opt = "GTE"
							//*
							gteTocNext, gteLitNext := p.scanIgnoreWhitespace()
							// fmt.Println("=> gte int value", gteTocNext, gteLitNext)
							if gteTocNext != IDENT {
								p.unscan()
								return nil, fmt.Errorf("found %q expecter int value", gteLitNext)
							}
							gteLitNextFloat, err := strconv.ParseFloat(gteLitNext, 64)
							if err != nil {
								panic(err)
							}
							//*
							O.Value = gteLitNextFloat
							O.ValueType = "Float64"
							//*
							I2O[conditionIndexName] = O
							I2OSet = append(I2OSet, I2O)
						}

					case LT:
						{
							//*
							O.Opt = "LT"
							//*
							ltTocNext, ltLitNext := p.scanIgnoreWhitespace()
							// fmt.Println("=> lt int value", ltTocNext, ltLitNext)
							if ltTocNext != IDENT {
								p.unscan()
								return nil, fmt.Errorf("found %q expecter int value", ltLitNext)
							}
							ltLitNextFloat, err := strconv.ParseFloat(ltLitNext, 64)
							if err != nil {
								panic(err)
							}
							O.Value = ltLitNextFloat
							O.ValueType = "Float64"
							I2O[conditionIndexName] = O
							I2OSet = append(I2OSet, I2O)
						}
					case LTE:
						{
							//*
							O.Opt = "LTE"
							//*
							lteTocNext, lteLitNext := p.scanIgnoreWhitespace()
							// fmt.Println("=> lt int value", lteTocNext, lteLitNext)
							if lteTocNext != IDENT {
								p.unscan()
								return nil, fmt.Errorf("found %q expecter int value", lteLitNext)
							}
							lteLitNextFloat, err := strconv.ParseFloat(lteLitNext, 64)
							if err != nil {
								panic(err)
							}
							O.Value = lteLitNextFloat
							O.ValueType = "Float64"
							I2O[conditionIndexName] = O
							I2OSet = append(I2OSet, I2O)
						}

					case PF:
						{
							//*
							O.Opt = "PF"
							//*
							pfTocNext, pfLitNext := p.scanIgnoreWhitespace()
							// fmt.Println("=> pf string value begin", pfTocNext, pfLitNext)
							if pfTocNext != STR {
								p.unscan()
								return nil, fmt.Errorf("found %q expecter string value prefix", pfLitNext)
							}
							pfTocNextT, pfLitNextL := p.scan()
							// fmt.Println("=> pf string value", pfTocNextT, pfLitNextL)
							if pfTocNextT != IDENT {
								p.unscan()
								return nil, fmt.Errorf("found %q expecter string value ", pfLitNextL)
							}
							pfTocEnd, pfLitEnd := p.scanIgnoreWhitespace()
							// fmt.Println("=> pf string value end", pfTocEnd, pfLitEnd)
							if pfTocEnd != STR {
								p.unscan()
								return nil, fmt.Errorf("found %q expecter string value end", pfLitEnd)
							}
							O.Value = pfLitNextL
							O.ValueType = "string"
							I2O[conditionIndexName] = O
							I2OSet = append(I2OSet, I2O)
						}
					case SF:
						{
							//*
							O.Opt = "SF"
							//*
							sfTocNext, sfLitNext := p.scanIgnoreWhitespace()
							// fmt.Println("=> sf string value begin", sfTocNext, sfLitNext)
							if sfTocNext != STR {
								p.unscan()
								return nil, fmt.Errorf("found %q expecter string value prefix", sfLitNext)
							}
							sfTocNextT, sfLitNextT := p.scan()
							// fmt.Println("=> sf string value", sfTocNextT, sfLitNextT)
							if sfTocNextT != IDENT {
								p.unscan()
								return nil, fmt.Errorf("found %q expecter string value", sfLitNextT)
							}
							sfTocEnd, sfLitEnd := p.scanIgnoreWhitespace()
							// fmt.Println("=> sf string value end", sfTocEnd, sfLitEnd)
							if sfTocEnd != STR {
								p.unscan()
								return nil, fmt.Errorf("found %q expecter string value end", sfLitEnd)
							}
							O.Value = sfLitNextT
							O.ValueType = "string"
							I2O[conditionIndexName] = O
							I2OSet = append(I2OSet, I2O)

						}
					case EQ:
						{
							//*
							O.Opt = "EQ"
							//*
							eqTocNext, eqLitNext := p.scanIgnoreWhitespace()
							if eqTocNext != STR && eqTocNext != IDENT {
								return nil, fmt.Errorf("found %q expect eq value", eqLitNext)
							}
							// fmt.Println("=> case eq", eqTocNext, eqLitNext)
							switch eqTocNext {
							case STR:
								{
									eqStrTocNext, eqStrLitNext := p.scan()
									// fmt.Println("=> eq string value", eqStrTocNext, eqStrLitNext)
									if eqStrTocNext != IDENT {
										p.unscan()
										return nil, fmt.Errorf("found %q expect: string value", eqStrLitNext)
									}
									eqStrTocNextL, eqStrLitNextL := p.scanIgnoreWhitespace()
									// fmt.Println("=> eq string value end", eqStrTocNextL, eqStrLitNextL)
									if eqStrTocNextL != STR {
										return nil, fmt.Errorf("found %q expect: string value end", eqStrLitNextL)
									}
									O.Value = eqStrLitNext
									O.ValueType = "string"
									I2O[conditionIndexName] = O
									I2OSet = append(I2OSet, I2O)

								}
							case IDENT:
								{
									floatValue, err := strconv.ParseFloat(eqLitNext, 64)
									if err != nil {
										panic(err)
									}
									O.Value = floatValue
									O.ValueType = "Float64"
									I2O[conditionIndexName] = O
									I2OSet = append(I2OSet, I2O)
								}

							}

						}
					case NEQ:
						{
							//*
							O.Opt = "NEQ"
							//*
							neqTocNext, neqLitNext := p.scanIgnoreWhitespace()
							if neqTocNext != STR && neqTocNext != IDENT {
								return nil, fmt.Errorf("found %q expect neq value.", neqLitNext)
							}
							// fmt.Println("=> neq case", neqTocNext, neqLitNext)
							switch neqTocNext {
							case STR:
								{
									neqTocNextL, neqLitNextL := p.scan()
									// fmt.Println("=> neq string value", neqLitNextL)
									if neqTocNextL != IDENT {
										return nil, fmt.Errorf("found %q expect neq string value", neqLitNextL)
									}
									neqTocNextEnd, neqLitNextEnd := p.scanIgnoreWhitespace()
									// fmt.Println("=> neq sting value end", neqTocNextEnd, neqLitNextEnd)
									if neqTocNextEnd != STR {
										return nil, fmt.Errorf("found %q expect neq string value end", neqLitNextEnd)
									}
									O.Value = neqLitNextL
									O.ValueType = "string"
									I2O[conditionIndexName] = O
									I2OSet = append(I2OSet, I2O)
								}
							case IDENT:
								{
									valueFloat, err := strconv.ParseFloat(neqLitNext, 64)
									if err != nil {
										return nil, fmt.Errorf("found %q expect value: type int", neqLitNext)
									}
									O.Value = valueFloat
									O.ValueType = "Float64"
									I2O[conditionIndexName] = O
									I2OSet = append(I2OSet, I2O)
								}
							}
						}
					}

					stmt.IndexToFieldSet = append(stmt.IndexToFieldSet, I2OSet...)
					toc, lit = p.scanIgnoreWhitespace()
					if toc != COMMA && toc != MParRight {
						p.unscan()
						return nil, fmt.Errorf("fount %q expecter", lit)
					}
					if toc == MParRight {
						stmt.IndexToFieldSet = append(stmt.IndexToFieldSet, IndexToFieldMap)
						// return stmt, nil
						break
					}
				}
			}
			atToken, atLit := p.scanIgnoreWhitespace()
			// fmt.Println("=> AT | ", atToken, atLit)
			if atToken != AT {
				p.unscan()
				return nil, fmt.Errorf("found %q expect AT", atLit)
			}
			atParLeftToken, atParLeftLit := p.scanIgnoreWhitespace()
			// fmt.Println("=> AT par | ", atParLeftToken, atParLeftLit)
			if atParLeftToken != MParLeft {
				p.unscan()
				return nil, fmt.Errorf("found %q expect [", atParLeftLit)
			}

			//*
			timeBeginStr := ""
			timeEndStr := ""
			//*
			timeTok, timeLit := p.scanIgnoreWhitespace()
			// fmt.Println("=> time value", timeTok, timeLit)
			if timeTok != IDENT {
				p.unscan()
				return nil, fmt.Errorf("found %q expect time value", timeLit)
			}
			//*
			timeBeginStr += timeLit
			//*
			timeTok, timeLit = p.scanIgnoreWhitespace()
			// fmt.Println("=> time sep", timeTok, timeLit)
			if timeTok != IS {
				p.unscan()
				return nil, fmt.Errorf("found %q expect .", timeLit)
			}
			//*
			timeBeginStr += ":"
			//*
			timeTok, timeLit = p.scanIgnoreWhitespace()
			// fmt.Println("=> time value", timeTok, timeLit)
			//*
			timeBeginStr += timeLit
			stmt.TimeBegin = timeBeginStr
			//*
			timeTok, timeLit = p.scanIgnoreWhitespace()
			// fmt.Println("=> mid", timeTok, timeLit)
			if timeTok != MIDEND {
				p.unscan()
				return nil, fmt.Errorf("found %q expect - ", timeLit)
			}
			timeTok, timeLit = p.scanIgnoreWhitespace()
			// fmt.Println("=> time value ", timeLit)
			if timeTok != IDENT {
				p.unscan()
				return nil, fmt.Errorf("found %q expect time value", timeLit)
			}
			//*
			timeEndStr += timeLit
			//*
			timeTok, timeLit = p.scanIgnoreWhitespace()
			if timeTok != IS {
				p.unscan()
				return nil, fmt.Errorf("found %q expect :", timeLit)
			}
			timeEndStr += ":"
			timeTok, timeLit = p.scanIgnoreWhitespace()
			// fmt.Println("=> time value", timeLit)
			if timeTok != IDENT {
				p.unscan()
				return nil, fmt.Errorf("found %q expect time value", timeLit)
			}
			//*
			timeEndStr += timeLit
			stmt.TimeEnd = timeEndStr
			//*
			timeTok, timeLit = p.scanIgnoreWhitespace()
			if timeTok != MParRight {
				p.unscan()
				return nil, fmt.Errorf("found %q expect ]", timeLit)
			} else {
				return stmt, nil
			}
			/*
				// fmt.Println("TIME:", timeBeginStr)
				//*
				stmt.TimeBegin = timeBeginStr
				//*
				timeTokEnd, timeLitEnd := p.scanIgnoreWhitespace()
				// fmt.Println("=> time end | ", timeTokEnd, timeLitEnd)
				if timeTokEnd != WS {
					p.unscan()
					return nil, fmt.Errorf("found %q expect time", timeLitEnd)
				}
				//*
				timeEndStr += timeLitEnd
				stmt.TimeEnd = timeEndStr
				//*
				timeTok, timeLit = p.scanIgnoreWhitespace()
				if timeTok != MParRight {
					p.unscan()
					return nil, fmt.Errorf("found %q expect ]", timeLit)
				} else {
					return stmt, nil
				}
			*/
		}

	}

	return stmt, nil
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }
