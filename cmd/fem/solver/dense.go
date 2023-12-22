package solver

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
	"wfem/cmd/fem/mesh"
	"wfem/cmd/fem/progress"
)

type DenseSolver struct {
	matrix mat.SymDense
	vector mat.VecDense
}

func NewDenseSolver(mesh *mesh.Mesh) *DenseSolver {
	return &DenseSolver{matrix: *mat.NewSymDense(mesh.NumVertex()*mesh.Freedom(), nil), vector: *mat.NewVecDense(mesh.NumVertex()*mesh.Freedom(), nil)}
}

func (ds *DenseSolver) SetMatrix(i, j int, value float64) {
	if i >= j {
		ds.matrix.SetSym(i, j, value)
	}
}

func (ds *DenseSolver) AddMatrix(i, j int, value float64) {
	if i >= j {
		ds.matrix.SetSym(i, j, ds.matrix.At(i, j)+value)
	}
}

func (ds *DenseSolver) SetVector(i int, value float64) {
	ds.vector.SetVec(i, value)
}

func (ds *DenseSolver) AddVector(i int, value float64) {
	ds.vector.SetVec(i, ds.vector.AtVec(i)+value)
}

func (ds *DenseSolver) GetMatrix(i, j int) float64 {
	return ds.matrix.At(i, j)
}

func (ds *DenseSolver) GetVector(i int) float64 {
	return ds.vector.AtVec(i)
}

func (ds *DenseSolver) SetBoundaryCondition(index int, value float64) {
	rows, _ := ds.matrix.Dims()
	for i := 0; i < rows; i++ {
		if i != index {
			if ds.matrix.At(index, i) != 0.0 {
				ds.matrix.SetSym(index, i, value)
				ds.matrix.SetSym(i, index, value)
			}
		}
	}
	ds.vector.SetVec(index, value*ds.matrix.At(index, index))
}

//func printMatrix(m *mat.SymDense, v *mat.VecDense) {
//	r, c := m.Dims()
//	for i := 0; i < r; i++ {
//		for j := 0; j < c; j++ {
//			print(m.At(i, j), "\t")
//		}
//		print(v.AtVec(i), "\t")
//		println("")
//	}
//}

//func saveMatrix(m *mat.SymDense, v *mat.VecDense) error {
//	file, err := os.Create("matr.txt")
//	if err != nil {
//		return fmt.Errorf("error creating file")
//	}
//	defer func() {
//		err = file.Close()
//	}()
//	r, c := m.Dims()
//	for i := 0; i < r; i++ {
//		for j := 0; j < c; j++ {
//			fmt.Fprintf(file, "%0.4e\t", m.At(i, j))
//		}
//		fmt.Fprintf(file, "%0.4e\n", v.AtVec(i))
//	}
//	return err
//}

//func saveMatrix(m *mat.SymDense, v *mat.VecDense) error {
//	file, err := os.Create("matr.txt")
//	if err != nil {
//		return fmt.Errorf("error creating file")
//	}
//	defer func() {
//		err = file.Close()
//	}()
//	r, _ := m.Dims()
//	for i := 0; i < r; i++ {
//		fmt.Fprintf(file, "%0.4e\n", v.AtVec(i))
//	}
//	return err
//}

// Solve example - https://github.com/gonum/gonum/blob/master/mat/cholesky_example_test.go
func (ds *DenseSolver) Solve() (*mat.VecDense, error) {
	var chl mat.Cholesky
	var x mat.VecDense
	// printMatrix(&ds.matrix, &ds.vector)
	// saveMatrix(&ds.matrix, &ds.vector)
	msg := progress.NewUnlimitedProgress("Solution of the system of equations")
	defer msg.StopProgress()
	if err := chl.Factorize(&ds.matrix); !err {
		return &ds.vector, fmt.Errorf("a matrix is not positive semi-definite")
	}
	if err := chl.SolveVecTo(&x, &ds.vector); err != nil {
		return nil, fmt.Errorf("matrix is near singular")
	}
	return &x, nil
}
