package fem

import (
	"gonum.org/v1/gonum/mat"
	"wfem/cmd/fem/mesh"
)

type FiniteElementMethod interface {
	SetMesh(string) error
	Calculate() error
	AddBoundaryCondition(string, string, int)
	AddVolumeLoad(string, string, int)
	AddSurfaceLoad(string, string, int)
	AddPointLoad(string, string, int)
	AddPressureLoad(string, string)
	AddThickness(string, string)
	AddPoissonRatio(string, string)
	AddYoungModulus(string, string)
	AddVariable(string, float64)
	SetNumTread(int)
	SetEps(float64)
	SaveResult(string) error
	GetResult() *mat.Dense
	GetResultNames() *[]string
	GetMesh() *mesh.Mesh
}
