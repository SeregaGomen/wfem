package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// Token type
const (
	Delimiter = iota + 1
	Numeric
	Variable
	Function
	Finished
)

type Value struct {
	name  string
	value float64
}

type Values []Value

func (v *Values) Set(name string, value float64) {
	for i := range *v {
		if (*v)[i].name == name {
			(*v)[i].value = value
			return
		}
	}
	*v = append(*v, Value{name, value})
}

func (v *Values) Get(name string) (*float64, bool) {
	for i := range *v {
		if (*v)[i].name == name {
			return &(*v)[i].value, true
		}
	}
	return nil, false
}

type Parser struct {
	result     Node
	variables  Values
	expression string
	token      string
	tok        int
	tokenType  int
}

func New() Parser {
	return Parser{result: Node{}, variables: Values{}}
}

func (p *Parser) SetExpression(exp string) error {
	p.expression = exp
	return p.compile()
}

func (p *Parser) SetVariable(name string, value float64) {
	p.variables.Set(name, value)
}

func (p *Parser) Value() (float64, error) {
	return p.result.Value()
}

func (p *Parser) compile() error {
	p.tok = 0
	p.tokenType = 0
	for {
		if p.tokenType == Finished {
			break
		}
		res, err := p.getExp()
		if err != nil {
			return err
		}
		p.result = *res
		//if p.tokenType == Delimiter {
		//	if p.tok == Rb {
		//		return fmt.Errorf("unbalanced brackets")
		//	} else {
		//		return fmt.Errorf("syntax error")
		//	}
		//}
	}
	return nil
}

func (p *Parser) findDelimiter(value string) bool {
	switch value {
	case "+":
		p.tok = Plus
	case "-":
		p.tok = Minus
	case "*":
		p.tok = Mul
	case "/":
		p.tok = Div
	case "**":
		p.tok = Pow
	case "or":
		p.tok = Or
	case "and":
		p.tok = And
	case "not":
		p.tok = Not
	case "==":
		p.tok = Eq
	case "!=":
		p.tok = Ne
	case "<=":
		p.tok = Le
	case "<":
		p.tok = Lt
	case ">=":
		p.tok = Ge
	case ">":
		p.tok = Gt
	case "(":
		p.tok = Lb
	case ")":
		p.tok = Rb
	default:
		return false
	}
	//p.tokenType = Delimiter
	return true
}

func (p *Parser) findFunction(value string) bool {
	switch strings.ToLower(value) {
	case "sqrt":
		p.tok = Sqrt
	case "sin":
		p.tok = Sin
	case "cos":
		p.tok = Cos
	case "tan":
		p.tok = Tan
	case "exp":
		p.tok = Exp
	case "asin":
		p.tok = Asin
	case "acos":
		p.tok = Acos
	case "atan":
		p.tok = Atan
	case "abs":
		p.tok = Abs
	default:
		return false
	}
	//p.tokenType = Function
	return true
}

func (p *Parser) getToken() error {
	p.token = ""
	p.tok = 0
	p.tokenType = 0
	// Skip leading spaces
	for len(p.expression) != 0 {
		if p.expression[0] != ' ' && p.expression[0] != '\t' {
			break
		}
		p.expression = p.expression[1:]
	}
	// Handling an empty string
	if len(p.expression) == 0 {
		p.tokenType = Finished
		return nil
	}
	// Delimiter handling
	if len(p.expression) != 0 && strings.Contains("+-*/()=^<>!", p.expression[0:1]) {
		p.token += p.expression[0:1]
		p.expression = p.expression[1:]
		// Check for double delimiter
		//if len(p.expression) != 0 && strings.Contains("=><", p.expression[0:1]) {
		if len(p.expression) != 0 && ((p.token+p.expression[0:1]) == "==" || (p.token+p.expression[0:1]) == "!=" ||
			(p.token+p.expression[0:1]) == "<=" || (p.token+p.expression[0:1]) == ">=" ||
			(p.token+p.expression[0:1]) == "**") {
			p.token += p.expression[0:1]
			p.expression = p.expression[1:]
		}
		if !p.findDelimiter(p.token) {
			return fmt.Errorf("syntax error")
		}
		p.tokenType = Delimiter
		return nil
	}
	// Handling the number
	if len(p.expression) != 0 && (p.expression[0] >= '0' && p.expression[0] <= '9') {
		for len(p.expression) != 0 && (p.expression[0] >= '0' && p.expression[0] <= '9') {
			p.token += p.expression[0:1]
			p.expression = p.expression[1:]
		}
		if len(p.expression) != 0 && p.expression[0] == '.' {
			p.token += "."
			p.expression = p.expression[1:]
			for len(p.expression) != 0 && (p.expression[0] >= '0' && p.expression[0] <= '9') {
				p.token += p.expression[0:1]
				p.expression = p.expression[1:]
			}
		}
		if len(p.expression) != 0 && (p.expression[0] == 'e' || p.expression[0] == 'E') {
			p.token += "E"
			p.expression = p.expression[1:]
			if len(p.expression) != 0 && (p.expression[0] == '+' || p.expression[0] == '-') {
				p.token += p.expression[0:1]
				p.expression = p.expression[1:]
				for len(p.expression) != 0 && (p.expression[0] >= '0' && p.expression[0] <= '9') {
					p.token += p.expression[0:1]
					p.expression = p.expression[1:]
				}
			} else {
				return fmt.Errorf("syntax error")
			}
		}
		p.tokenType = Numeric
		return nil
	}
	// Process string literal
	if len(p.expression) != 0 && ((p.expression[0] >= 'a' && p.expression[0] <= 'z') ||
		(p.expression[0] >= 'A' && p.expression[0] <= 'Z') || p.expression[0] == '_') {
		for len(p.expression) != 0 && ((p.expression[0] >= 'a' && p.expression[0] <= 'z') ||
			(p.expression[0] >= 'A' && p.expression[0] <= 'Z') || p.expression[0] == '_' ||
			(p.expression[0] >= '0' && p.expression[0] <= '9')) {
			p.token += p.expression[0:1]
			p.expression = p.expression[1:]
		}
		if p.findDelimiter(p.token) {
			p.tokenType = Delimiter
			return nil
		} else if p.findFunction(p.token) {
			p.tokenType = Function
			return nil
		}
		p.tokenType = Variable
		return nil
	}
	return fmt.Errorf("syntax error")
}

func (p *Parser) getExp() (*Node, error) {
	if err := p.getToken(); err != nil {
		return nil, err
	}
	return p.tokenOr()
}

func (p *Parser) tokenOr() (*Node, error) {
	res, err := p.tokenAnd()
	if err != nil {
		return nil, err
	}
	for p.tokenType != Finished && p.tok == Or {
		if err = p.getToken(); err != nil {
			return nil, err
		}
		hold, err := p.tokenAnd()
		if err != nil {
			return nil, err
		}
		*res = NewBinary(*res, Or, *hold)
	}
	return res, nil
}

func (p *Parser) tokenAnd() (*Node, error) {
	res, err := p.tokenNot()
	if err != nil {
		return nil, err
	}
	for p.tokenType != Finished && p.tok == And {
		if err = p.getToken(); err != nil {
			return nil, err
		}
		hold, err := p.tokenNot()
		if err != nil {
			return nil, err
		}
		*res = NewBinary(*res, And, *hold)
	}
	return res, nil
}

func (p *Parser) tokenNot() (*Node, error) {
	op := p.tok
	if p.tokenType != Finished && p.tok == Not {
		if err := p.getToken(); err != nil {
			return nil, err
		}
	}
	res, err := p.tokenEq()
	if err != nil {
		return nil, err
	}
	if op == Not {
		*res = NewUnary(Not, *res)
	}
	return res, nil
}

func (p *Parser) tokenEq() (*Node, error) {
	res, err := p.tokenAdd()
	if err != nil {
		return nil, err
	}
	for p.tokenType != Finished && (p.tok == Gt || p.tok == Ge || p.tok == Lt || p.tok == Le || p.tok == Eq ||
		p.tok == Ne) {
		op := p.tok
		if err = p.getToken(); err != nil {
			break
		}
		hold, err := p.tokenAdd()
		if err != nil {
			break
		}
		*res = NewBinary(*res, op, *hold)
	}
	return res, err
}

func (p *Parser) tokenAdd() (*Node, error) {
	res, err := p.tokenMul()
	if err != nil {
		return nil, err
	}
	for p.tokenType != Finished && (p.tok == Plus || p.tok == Minus) {
		op := p.tok
		if err = p.getToken(); err != nil {
			break
		}
		hold, err := p.tokenMul()
		if err != nil {
			break
		}
		*res = NewBinary(*res, op, *hold)
	}
	return res, err
}

func (p *Parser) tokenMul() (*Node, error) {
	res, err := p.tokenPow()
	if err != nil {
		return nil, err
	}
	for p.tokenType != Finished && (p.tok == Mul || p.tok == Div) {
		op := p.tok
		if err = p.getToken(); err != nil {
			break
		}
		hold, err := p.tokenPow()
		if err != nil {
			break
		}
		*res = NewBinary(*res, op, *hold)
	}
	return res, err
}

func (p *Parser) tokenPow() (*Node, error) {
	res, err := p.tokenUn()
	if err != nil {
		return nil, err
	}
	for p.tokenType != Finished && p.tok == Pow {
		if err = p.getToken(); err != nil {
			break
		}
		hold, err := p.tokenBracket()
		if err != nil {
			break
		}
		*res = NewBinary(*res, Pow, *hold)
	}
	return res, err
}

func (p *Parser) tokenUn() (*Node, error) {
	var op int
	if p.tok == Plus || p.tok == Minus {
		op = p.tok
	}
	if p.tokenType == Delimiter && (p.tok == Plus || p.tok == Minus) {
		if err := p.getToken(); err != nil {
			return nil, err
		}
	}
	res, err := p.tokenBracket()
	if err != nil {
		return nil, err
	}
	if op == Plus || op == Minus {
		*res = NewUnary(op, *res)
	}
	return res, err
}

func (p *Parser) tokenBracket() (*Node, error) {
	var res *Node
	var err error
	if p.tokenType == Delimiter && p.tok == Lb {
		if err = p.getToken(); err != nil {
			return nil, err
		}
		res, err = p.tokenOr()
		if err != nil {
			return nil, err
		}
		if p.tok != Rb {
			return nil, fmt.Errorf("unbalanced brackets")
		}
		if err = p.getToken(); err != nil {
			return nil, err
		}
	} else {
		res, err = p.tokenPrim()
		if err != nil {
			return nil, err
		}
	}
	return res, err
}

func (p *Parser) tokenPrim() (*Node, error) {
	var res Node
	switch p.tokenType {
	case Numeric:
		val, err := strconv.ParseFloat(p.token, 64)
		if err != nil {
			return nil, err
		}
		res = NewDouble(&val)
		if err = p.getToken(); err != nil {
			return nil, err
		}
	case Variable:
		if val, ok := p.variables.Get(p.token); ok {
			res = NewDouble(val)
			if err := p.getToken(); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("undefined variable")
		}
	case Function:
		return p.tokenFunction()
	default:
		return nil, fmt.Errorf("syntax error")
	}
	return &res, nil
}

func (p *Parser) tokenFunction() (*Node, error) {
	var err error
	var res *Node
	funTok := p.tok
	if err = p.getToken(); err != nil {
		return nil, err
	}
	if len(p.token) == 0 || p.tok != Lb {
		return nil, fmt.Errorf("syntax error")
	}
	if err = p.getToken(); err != nil {
		return nil, err
	}
	if res, err = p.tokenAdd(); err != nil {
		return nil, err
	}
	*res = NewUnary(funTok, *res)
	if p.tok != Rb {
		return res, fmt.Errorf("syntax error")
	}
	if err = p.getToken(); err != nil {
		return res, err
	}
	return res, err
}
