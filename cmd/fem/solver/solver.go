package solver

import (
	"gonum.org/v1/gonum/mat"
)

type Solver interface {
	SetMatrix(int, int, float64)
	AddMatrix(int, int, float64)
	SetBoundaryCondition(int, float64)
	SetVector(int, float64)
	AddVector(int, float64)
	GetMatrix(int, int) float64
	GetVector(int) float64
	Solve() (*mat.VecDense, error)
}
