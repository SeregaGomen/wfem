package parser

import (
	"fmt"
	"math"
)

// Token
const (
	Number = iota + 1
	Sin
	Cos
	Tan
	Exp
	Asin
	Acos
	Atan
	Sqrt
	Abs
	Plus
	Minus
	Div
	Mul
	Pow
	Or
	And
	Not
	Eq
	Ne
	Le
	Lt
	Ge
	Gt
	Lb
	Rb
)

type Node struct {
	value    *float64
	token    int
	children []Node
}

func NewDouble(val *float64) Node {
	return Node{value: val, token: Number, children: []Node{}}
}

func NewUnary(token int, rhs Node) Node {
	return Node{token: token, children: []Node{rhs}}
}

func NewBinary(lhs Node, token int, rhs Node) Node {
	return Node{token: token, children: []Node{lhs, rhs}}
}

func (n Node) Value() (float64, error) {
	var (
		err error
		val float64
		lhs float64
		rhs float64
	)

	switch n.token {
	case Number:
		val = *n.value
	case Plus:
		if len(n.children) == 1 {
			val, err = n.children[0].Value()
		} else {
			lhs, err = n.children[0].Value()
			if err == nil {
				rhs, err = n.children[1].Value()
				if err == nil {
					val = lhs + rhs
				}
			}
		}
	case Minus:
		if len(n.children) == 1 {
			val, err = n.children[0].Value()
			val = -val
		} else {
			lhs, err = n.children[0].Value()
			if err == nil {
				rhs, err = n.children[1].Value()
				if err == nil {
					val = lhs - rhs
				}
			}
		}
	case Mul:
		lhs, err = n.children[0].Value()
		if err == nil {
			rhs, err = n.children[1].Value()
			if err == nil {
				val = lhs * rhs
			}
		}
	case Div:
		lhs, err = n.children[0].Value()
		if err == nil {
			rhs, err = n.children[1].Value()
			if err == nil && rhs != 0 {
				val = lhs / rhs
			} else if err == nil {
				err = fmt.Errorf("devide by zero")
			}
		}
	case Pow:
		lhs, err = n.children[0].Value()
		if err == nil {
			rhs, err = n.children[1].Value()
			if err == nil {
				val = math.Pow(lhs, rhs)
			}
		}
	case Sin:
		val, err = n.children[0].Value()
		if err == nil {
			val = math.Sin(val)
		}
	case Cos:
		val, err = n.children[0].Value()
		if err == nil {
			val = math.Cos(val)
		}
	case Tan:
		val, err = n.children[0].Value()
		if err == nil {
			val = math.Tan(val)
		}
	case Abs:
		val, err = n.children[0].Value()
		if err == nil {
			val = math.Abs(val)
		}
	case Exp:
		val, err = n.children[0].Value()
		if err == nil {
			val = math.Exp(val)
		}
	case Asin:
		val, err = n.children[0].Value()
		if err == nil {
			val = math.Asin(val)
		}
	case Acos:
		val, err = n.children[0].Value()
		if err == nil {
			val = math.Acos(val)
		}
	case Atan:
		val, err = n.children[0].Value()
		if err == nil {
			val = math.Atan(val)
		}
	case Sqrt:
		val, err = n.children[0].Value()
		if err == nil {
			if val < 0 {
				err = fmt.Errorf("square root of a negative number")
			} else {
				val = math.Sqrt(val)
			}
		}
	case And:
		lhs, err = n.children[0].Value()
		if err == nil {
			rhs, err = n.children[1].Value()
			if err == nil {
				if lhs == 1.0 && rhs == 1.0 {
					val = 1.0
				} else {
					val = 0.0
				}
			}
		}
	case Or:
		lhs, err = n.children[0].Value()
		if err == nil {
			rhs, err = n.children[1].Value()
			if err == nil {
				if lhs == 1.0 || rhs == 1.0 {
					val = 1.0
				} else {
					val = 0.0
				}
			}
		}
	case Not:
		val, err = n.children[0].Value()
		if err == nil {
			if val == 1.0 {
				val = 0.0
			} else {
				val = 1.0
			}
		}
	case Eq:
		lhs, err = n.children[0].Value()
		if err == nil {
			rhs, err = n.children[1].Value()
			if err == nil {
				if lhs == rhs {
					val = 1.0
				} else {
					val = 0.0
				}
			}
		}
	case Ne:
		lhs, err = n.children[0].Value()
		if err == nil {
			rhs, err = n.children[1].Value()
			if err == nil {
				if lhs != rhs {
					val = 1.0
				} else {
					val = 0.0
				}
			}
		}
	case Le:
		lhs, err = n.children[0].Value()
		if err == nil {
			rhs, err = n.children[1].Value()
			if err == nil {
				if lhs <= rhs {
					val = 1.0
				} else {
					val = 0.0
				}
			}
		}
	case Lt:
		lhs, err = n.children[0].Value()
		if err == nil {
			rhs, err = n.children[1].Value()
			if err == nil {
				if lhs < rhs {
					val = 1.0
				} else {
					val = 0.0
				}
			}
		}
	case Ge:
		lhs, err = n.children[0].Value()
		if err == nil {
			rhs, err = n.children[1].Value()
			if err == nil {
				if lhs >= rhs {
					val = 1.0
				} else {
					val = 0.0
				}
			}
		}
	case Gt:
		lhs, err = n.children[0].Value()
		if err == nil {
			rhs, err = n.children[1].Value()
			if err == nil {
				if lhs > rhs {
					val = 1.0
				} else {
					val = 0.0
				}
			}
		}
	default:
		err = fmt.Errorf("unknown operation")
	}
	return val, err
}
