package params

import (
	"gonum.org/v1/gonum/mat"
	"wfem/cmd/parser"
)

// Direct
const (
	X = 1
	Y = 2
	Z = 4
)

// Type of parameters
const (
	BoundaryCondition int = iota
	VolumeLoad
	SurfaceLoad
	PointLoad
	PressureLoad
	Thickness
	YoungModulus
	PoissonRatio
)

type Parameter struct {
	Type      int
	Value     string
	Predicate string
	Direct    int
}

func (p Parameter) GetValue(x *mat.VecDense, variables *map[string]float64) (float64, error) {
	var value float64
	var err error
	valueParser := parser.New()
	for name, value := range *variables {
		valueParser.SetVariable(name, value)
	}
	valueParser.SetVariable("x", x.AtVec(0))
	if rows, _ := x.Dims(); rows > 1 {
		valueParser.SetVariable("y", x.AtVec(1))
	}
	if rows, _ := x.Dims(); rows > 2 {
		valueParser.SetVariable("z", x.AtVec(2))
	}
	if err = valueParser.SetExpression(p.Value); err != nil {
		return value, err
	}
	return valueParser.Value()
}

func (p Parameter) GetPredicate(x *mat.VecDense, variables *map[string]float64) (bool, error) {
	var value float64
	var err error
	if len(p.Predicate) == 0 {
		return true, nil
	}
	predicateParser := parser.New()
	for name, value := range *variables {
		predicateParser.SetVariable(name, value)
	}
	predicateParser.SetVariable("x", x.AtVec(0))

	if rows, _ := x.Dims(); rows > 1 {
		predicateParser.SetVariable("y", x.AtVec(1))
	}
	if rows, _ := x.Dims(); rows > 2 {
		predicateParser.SetVariable("z", x.AtVec(2))
	}
	if err = predicateParser.SetExpression(p.Predicate); err != nil {
		return false, err
	}
	if value, err = predicateParser.Value(); err != nil {
		return false, err
	}
	if value == 0.0 {
		return false, nil
	}
	return true, nil
}

type FEMParameters struct {
	Params    []Parameter
	Eps       float64
	NumThread int
	Variables map[string]float64
}

func New() FEMParameters {
	return FEMParameters{Params: []Parameter{}, Eps: 1.0e-10, NumThread: 1, Variables: map[string]float64{}}
}

func (p *FEMParameters) SetNumThread(n int) {
	p.NumThread = n
}

func (p *FEMParameters) SetEps(eps float64) {
	p.Eps = eps
}

func (p *FEMParameters) AddVariable(name string, value float64) {
	p.Variables[name] = value
}

func (p *FEMParameters) AddYoungModulus(value, predicate string) {
	p.Params = append(p.Params, Parameter{Type: YoungModulus, Value: value, Predicate: predicate})
}

func (p *FEMParameters) AddPoissonRatio(value, predicate string) {
	p.Params = append(p.Params, Parameter{Type: PoissonRatio, Value: value, Predicate: predicate})
}

func (p *FEMParameters) AddThickness(value, predicate string) {
	p.Params = append(p.Params, Parameter{Type: Thickness, Value: value, Predicate: predicate})
}

func (p *FEMParameters) AddPressureLoad(value, predicate string) {
	p.Params = append(p.Params, Parameter{Type: PressureLoad, Value: value, Predicate: predicate, Direct: X | Y | Z})
}

func (p *FEMParameters) AddConcentratedLoad(value, predicate string, direct int) {
	p.Params = append(p.Params, Parameter{Type: PointLoad, Value: value, Predicate: predicate, Direct: direct})
}

func (p *FEMParameters) AddSurfaceLoad(value, predicate string, direct int) {
	p.Params = append(p.Params, Parameter{Type: SurfaceLoad, Value: value, Predicate: predicate, Direct: direct})
}

func (p *FEMParameters) AddVolumeLoad(value, predicate string, direct int) {
	p.Params = append(p.Params, Parameter{Type: VolumeLoad, Value: value, Predicate: predicate, Direct: direct})
}

func (p *FEMParameters) AddBoundaryCondition(value, predicate string, direct int) {
	p.Params = append(p.Params, Parameter{Type: BoundaryCondition, Value: value, Predicate: predicate, Direct: direct})
}

func (p *FEMParameters) GetParamValue(x *mat.VecDense, pType int) (float64, error) {
	for i := range p.Params {
		if p.Params[i].Type == pType {
			isOk, err := p.Params[i].GetPredicate(x, &p.Variables)
			if err != nil {
				return 0, err
			}
			if isOk {
				return p.Params[i].GetValue(x, &p.Variables)
			}
		}
	}
	return 0, nil
}

func (p *FEMParameters) FindParameter(pType int) bool {
	for i := range p.Params {
		if p.Params[i].Type == pType {
			return true
		}
	}
	return false
}
