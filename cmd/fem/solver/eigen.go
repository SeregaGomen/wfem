package solver

/*
//#cgo CXXFLAGS: -std=c++20 -I../../../../../eigen
#cgo CXXFLAGS: -std=c++20 -I../../../../../eigen -I/usr/include/mkl
#cgo LDFLAGS: -L/lib/x86_64-linux-gnu -lmkl_intel_lp64 -lmkl_sequential -lmkl_core
#include "eigen.h"
*/

import (
	"C"
	"fmt"
	"unsafe"
	"wfem/cmd/fem/mesh"
	"wfem/cmd/fem/progress"

	"gonum.org/v1/gonum/mat"
)

type EigenSolver struct {
	size int
}

func NewEigenSolver(mesh *mesh.Mesh) *EigenSolver {
	maxNonZero := 0
	for i := range mesh.MeshMap {
		if len(mesh.MeshMap[i]) > maxNonZero {
			maxNonZero = len(mesh.MeshMap[i])
		}
	}
	C.InitMatrix((C.int)(mesh.NumVertex()*mesh.Freedom()), (C.int)(2*maxNonZero*mesh.Freedom()))
	return &EigenSolver{size: mesh.NumVertex() * mesh.Freedom()}
}

func (_ *EigenSolver) SetMatrix(i, j int, value float64) {
	C.SetMatrix((C.int)(i), (C.int)(j), (C.double)(value))
}

func (_ *EigenSolver) AddMatrix(i, j int, value float64) {
	C.AddMatrix((C.int)(i), (C.int)(j), (C.double)(value))
}

func (_ *EigenSolver) SetVector(i int, value float64) {
	C.SetVector((C.int)(i), (C.double)(value))
}

func (_ *EigenSolver) AddVector(i int, value float64) {
	C.AddVector((C.int)(i), (C.double)(value))
}

func (_ *EigenSolver) GetMatrix(i, j int) float64 {
	return float64(C.GetMatrix((C.int)(i), (C.int)(j)))
}

func (es *EigenSolver) GetVector(i int) float64 {
	return float64(C.GetVector((C.int)(i)))
}

func (es *EigenSolver) SetBoundaryCondition(i int, value float64) {
	C.SetBoundaryCondition((C.int)(i), (C.double)(value))
}

func (es *EigenSolver) Solve() (*mat.VecDense, error) {
	var err error
	x := make([]float64, es.size)
	msg := progress.NewUnlimitedProgress("Solution of the system of equations")
	defer msg.StopProgress()
	result := C.SolveEigen((*C.double)(unsafe.Pointer(&x[0])))
	if result != 0 {
		err = fmt.Errorf("matrix is near singular")
	}
	return mat.NewVecDense(es.size, x), err
}
