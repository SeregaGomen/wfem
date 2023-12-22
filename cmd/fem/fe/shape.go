package fe

import (
	"gonum.org/v1/gonum/mat"
	"wfem/cmd/fem/util"
)

type ShapeFunction1D interface {
	Size() int
	Shape(int, int) float64
	ShapeDxi(int, int) float64
	ShapeDx(int, int) float64
	W() *[]float64
	X() *mat.Dense
}

type ShapeFunction2D interface {
	ShapeFunction1D
	ShapeDeta(int, int) float64
	ShapeDy(int, int) float64
}

type ShapeFunction3D interface {
	ShapeFunction2D
	ShapeDpsi(int, int) float64
	ShapeDz(int, int) float64
}

type ShapeData struct {
	x   *mat.Dense
	c   *mat.Dense
	xi  *[]float64
	eta *[]float64
	psi *[]float64
	w   *[]float64
}

func (s *ShapeData) Size() int {
	rows, _ := s.c.Dims()
	return rows
}

func (s *ShapeData) W() *[]float64 {
	return s.w
}

func (s *ShapeData) X() *mat.Dense {
	return s.x
}

type Shape1d2 struct {
	ShapeData
}

func NewShape1d2(x *mat.Dense) (*Shape1d2, error) {
	c, err := util.CreateShape(2, func(i, j int) float64 { return [2]float64{1.0, x.At(i, 0)}[j] })
	if err != nil {
		return nil, err
	}
	return &Shape1d2{ShapeData{x: x, c: c, xi: &[]float64{-0.57735026919, 0.0, 0.57735026919},
		w: &[]float64{0.55555555556, 0.88888888889, 0.55555555556}}}, nil
}

func (s *Shape1d2) Shape(i, j int) float64 {
	return [2]float64{(1.0 - (*s.xi)[i]) * 0.5, (1.0 + (*s.xi)[i]) * 0.5}[j]
}

func (s *Shape1d2) ShapeDxi(_, j int) float64 {
	return [2]float64{-0.5, 0.5}[j]
}

func (s *Shape1d2) ShapeDx(_, j int) float64 {
	return s.c.At(1, j)
}

type Shape2d3 struct {
	ShapeData
}

func NewShape2d3(x *mat.Dense) (*Shape2d3, error) {
	c, err := util.CreateShape(3, func(i, j int) float64 { return [3]float64{1.0, x.At(i, 0), x.At(i, 1)}[j] })
	if err != nil {
		return nil, err
	}
	return &Shape2d3{ShapeData{x: x, c: c, xi: &[]float64{0.0, 0.5, 0.5},
		eta: &[]float64{0.5, 0.0, 0.5},
		w:   &[]float64{0.166666666667, 0.166666666667, 0.166666666667}}}, nil
}

func (s *Shape2d3) Shape(i, j int) float64 {
	return [3]float64{1.0 - (*s.xi)[i] - (*s.eta)[i], (*s.xi)[i], (*s.eta)[i]}[j]
}

func (s *Shape2d3) ShapeDxi(_, j int) float64 {
	return [3]float64{-1.0, 1.0, 0.0}[j]
}

func (s *Shape2d3) ShapeDeta(_, j int) float64 {
	return [3]float64{-1.0, 0.0, 1.0}[j]
}

func (s *Shape2d3) ShapeDx(_, j int) float64 {
	return s.c.At(1, j)
}

func (s *Shape2d3) ShapeDy(_, j int) float64 {
	return s.c.At(2, j)
}

type Shape2d4 struct {
	ShapeData
}

func NewShape2d4(x *mat.Dense) (*Shape2d4, error) {
	c, err := util.CreateShape(4, func(i, j int) float64 { return [4]float64{1.0, x.At(i, 0), x.At(i, 1), x.At(i, 0) * x.At(i, 1)}[j] })
	if err != nil {
		return nil, err
	}
	return &Shape2d4{ShapeData{x: x, c: c, xi: &[]float64{-0.57735027, -0.57735027, 0.57735027, 0.57735027},
		eta: &[]float64{-0.57735027, 0.57735027, -0.57735027, 0.57735027},
		w:   &[]float64{1.0, 1.0, 1.0, 1.0}}}, nil
}

func (s *Shape2d4) Shape(i, j int) float64 {
	return [4]float64{0.25 * (1.0 - (*s.xi)[i]) * (1.0 - (*s.eta)[i]), 0.25 * (1.0 + (*s.xi)[i]) * (1.0 - (*s.eta)[i]), 0.25 * (1.0 + (*s.xi)[i]) * (1.0 + (*s.eta)[i]), 0.25 * (1.0 - (*s.xi)[i]) * (1.0 + (*s.eta)[i])}[j]
}

func (s *Shape2d4) ShapeDxi(i, j int) float64 {
	return [4]float64{-0.25 * (1.0 - (*s.eta)[i]), 0.25 * (1.0 - (*s.eta)[i]), 0.25 * (1.0 + (*s.eta)[i]), -0.25 * (1.0 + (*s.eta)[i])}[j]
}

func (s *Shape2d4) ShapeDeta(i, j int) float64 {
	return [4]float64{-0.25 * (1.0 - (*s.xi)[i]), -0.25 * (1.0 + (*s.xi)[i]), 0.25 * (1.0 + (*s.xi)[i]), 0.25 * (1.0 - (*s.xi)[i])}[j]
}

func (s *Shape2d4) ShapeDx(i, j int) float64 {
	return s.c.At(1, j) + s.c.At(3, j)*s.x.At(i, 1)
}

func (s *Shape2d4) ShapeDy(i, j int) float64 {
	return s.c.At(2, j) + s.c.At(3, j)*s.x.At(i, 0)
}

type Shape3d4 struct {
	ShapeData
}

func NewShape3d4(x *mat.Dense) (*Shape3d4, error) {
	c, err := util.CreateShape(4, func(i, j int) float64 { return [4]float64{1.0, x.At(i, 0), x.At(i, 1), x.At(i, 2)}[j] })
	if err != nil {
		return nil, err
	}
	return &Shape3d4{ShapeData{x: x, c: c, xi: &[]float64{0.25, 0.5, 0.16666666667, 0.16666666667, 0.16666666667},
		eta: &[]float64{0.25, 0.16666666667, 0.5, 0.16666666667, 0.16666666667},
		psi: &[]float64{0.25, 0.16666666667, 0.16666666667, 0.5, 0.16666666667},
		w:   &[]float64{-0.13333333333, 0.075, 0.075, 0.075, 0.075}}}, nil
}

func (s *Shape3d4) Shape(i, j int) float64 {
	return [4]float64{1.0 - (*s.xi)[i] - (*s.eta)[i] - (*s.psi)[i], (*s.xi)[i], (*s.eta)[i], (*s.psi)[i]}[j]
}

func (s *Shape3d4) ShapeDxi(_, j int) float64 {
	return [4]float64{-1.0, 1.0, 0.0, 0.0}[j]
}

func (s *Shape3d4) ShapeDeta(_, j int) float64 {
	return [4]float64{-1.0, 0.0, 1.0, 0.0}[j]
}

func (s *Shape3d4) ShapeDpsi(_, j int) float64 {
	return [4]float64{-1.0, 0.0, 0.0, 1.0}[j]
}

func (s *Shape3d4) ShapeDx(_, j int) float64 {
	return s.c.At(1, j)
}

func (s *Shape3d4) ShapeDy(_, j int) float64 {
	return s.c.At(2, j)
}

func (s *Shape3d4) ShapeDz(_, j int) float64 {
	return s.c.At(3, j)
}

type Shape3d8 struct {
	ShapeData
}

func NewShape3d8(x *mat.Dense) (*Shape3d8, error) {
	c, err := util.CreateShape(8, func(i, j int) float64 {
		return [8]float64{1.0, x.At(i, 0), x.At(i, 1), x.At(i, 2), x.At(i, 0) * x.At(i, 1), x.At(i, 0) * x.At(i, 2), x.At(i, 1) * x.At(i, 2), x.At(i, 0) * x.At(i, 1) * x.At(i, 2)}[j]
	})
	if err != nil {
		return nil, err
	}
	return &Shape3d8{ShapeData{x: x, c: c, xi: &[]float64{-0.57735026919, -0.57735026919, -0.57735026919, -0.57735026919, 0.57735026919, 0.57735026919, 0.57735026919, 0.57735026919},
		eta: &[]float64{-0.57735026919, -0.57735026919, 0.57735026919, 0.57735026919, -0.57735026919, -0.57735026919, 0.57735026919, 0.57735026919},
		psi: &[]float64{-0.57735026919, 0.57735026919, -0.57735026919, 0.57735026919, -0.57735026919, 0.57735026919, -0.57735026919, 0.57735026919},
		w:   &[]float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0}}}, nil
}

func (s *Shape3d8) Shape(i, j int) float64 {
	return [8]float64{
		0.125 * (1.0 - (*s.xi)[i]) * (1.0 - (*s.eta)[i]) * (1.0 - (*s.psi)[i]),
		0.125 * (1.0 + (*s.xi)[i]) * (1.0 - (*s.eta)[i]) * (1.0 - (*s.psi)[i]),
		0.125 * (1.0 + (*s.xi)[i]) * (1.0 + (*s.eta)[i]) * (1.0 - (*s.psi)[i]),
		0.125 * (1.0 - (*s.xi)[i]) * (1.0 + (*s.eta)[i]) * (1.0 - (*s.psi)[i]),
		0.125 * (1.0 - (*s.xi)[i]) * (1.0 - (*s.eta)[i]) * (1.0 + (*s.psi)[i]),
		0.125 * (1.0 + (*s.xi)[i]) * (1.0 - (*s.eta)[i]) * (1.0 + (*s.psi)[i]),
		0.125 * (1.0 + (*s.xi)[i]) * (1.0 + (*s.eta)[i]) * (1.0 + (*s.psi)[i]),
		0.125 * (1.0 - (*s.xi)[i]) * (1.0 + (*s.eta)[i]) * (1.0 + (*s.psi)[i]),
	}[j]
}

func (s *Shape3d8) ShapeDxi(i, j int) float64 {
	return [8]float64{
		-0.125 * (1.0 - (*s.eta)[i]) * (1.0 - (*s.psi)[i]),
		0.125 * (1.0 - (*s.eta)[i]) * (1.0 - (*s.psi)[i]),
		0.125 * (1.0 + (*s.eta)[i]) * (1.0 - (*s.psi)[i]),
		-0.125 * (1.0 + (*s.eta)[i]) * (1.0 - (*s.psi)[i]),
		-0.125 * (1.0 - (*s.eta)[i]) * (1.0 + (*s.psi)[i]),
		0.125 * (1.0 - (*s.eta)[i]) * (1.0 + (*s.psi)[i]),
		0.125 * (1.0 + (*s.eta)[i]) * (1.0 + (*s.psi)[i]),
		-0.125 * (1.0 + (*s.eta)[i]) * (1.0 + (*s.psi)[i]),
	}[j]
}

func (s *Shape3d8) ShapeDeta(i, j int) float64 {
	return [8]float64{
		-0.125 * (1.0 - (*s.xi)[i]) * (1.0 - (*s.psi)[i]),
		-0.125 * (1.0 + (*s.xi)[i]) * (1.0 - (*s.psi)[i]),
		0.125 * (1.0 + (*s.xi)[i]) * (1.0 - (*s.psi)[i]),
		0.125 * (1.0 - (*s.xi)[i]) * (1.0 - (*s.psi)[i]),
		-0.125 * (1.0 - (*s.xi)[i]) * (1.0 + (*s.psi)[i]),
		-0.125 * (1.0 + (*s.xi)[i]) * (1.0 + (*s.psi)[i]),
		0.125 * (1.0 + (*s.xi)[i]) * (1.0 + (*s.psi)[i]),
		0.125 * (1.0 - (*s.xi)[i]) * (1.0 + (*s.psi)[i]),
	}[j]
}

func (s *Shape3d8) ShapeDpsi(i, j int) float64 {
	return [8]float64{
		-0.125 * (1.0 - (*s.xi)[i]) * (1.0 - (*s.eta)[i]),
		-0.125 * (1.0 + (*s.xi)[i]) * (1.0 - (*s.eta)[i]),
		-0.125 * (1.0 + (*s.xi)[i]) * (1.0 + (*s.eta)[i]),
		-0.125 * (1.0 - (*s.xi)[i]) * (1.0 + (*s.eta)[i]),
		0.125 * (1.0 - (*s.xi)[i]) * (1.0 - (*s.eta)[i]),
		0.125 * (1.0 + (*s.xi)[i]) * (1.0 - (*s.eta)[i]),
		0.125 * (1.0 + (*s.xi)[i]) * (1.0 + (*s.eta)[i]),
		0.125 * (1.0 - (*s.xi)[i]) * (1.0 + (*s.eta)[i]),
	}[j]
}

func (s *Shape3d8) ShapeDx(i, j int) float64 {
	return s.c.At(1, j) + s.c.At(4, j)*s.x.At(i, 1) + s.c.At(5, j)*s.x.At(i, 2) + s.c.At(7, j)*s.x.At(i, 1)*s.x.At(i, 2)
}

func (s *Shape3d8) ShapeDy(i, j int) float64 {
	return s.c.At(2, j) + s.c.At(4, j)*s.x.At(i, 0) + s.c.At(6, j)*s.x.At(i, 2) + s.c.At(7, j)*s.x.At(i, 0)*s.x.At(i, 2)
}

func (s *Shape3d8) ShapeDz(i, j int) float64 {
	return s.c.At(3, j) + s.c.At(5, j)*s.x.At(i, 0) + s.c.At(6, j)*s.x.At(i, 1) + s.c.At(7, j)*s.x.At(i, 0)*s.x.At(i, 1)
}
