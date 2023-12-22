package util

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
	"math"
)

func Gauss(a *mat.Dense, b *mat.VecDense, eps float64) bool {
	size, _ := a.Dims()
	// fmt.Printf("\rFactorization matrix: ")
	for i := 0; i < size-1; i++ {
		//if i%10 == 0 {
		//	fmt.Printf("\rMatrix factorization: %d%%", 100*i/(size-1))
		//}
		if math.Abs(a.At(i, i)) < eps {
			for j := i + 1; j < size; j++ {
				if math.Abs(a.At(j, i)) > eps {
					for k := 0; k < size; k++ {
						tmp := a.At(j, k)
						a.Set(j, k, a.At(i, k))
						a.Set(i, k, tmp)
					}
					tmp := b.AtVec(j)
					b.SetVec(j, b.AtVec(i))
					b.SetVec(i, tmp)
				}
			}
		}
		factor1 := a.At(i, i)
		for j := i + 1; j < size; j++ {
			factor2 := a.At(j, i)
			if math.Abs(factor2) > eps {
				for k := i; k < size; k++ {
					a.Set(j, k, a.At(j, k)-factor2*a.At(i, k)/factor1)
				}
				b.SetVec(j, b.AtVec(j)-factor2*b.AtVec(i)/factor1)
			}
		}
	}
	if math.Abs(a.At(size-1, size-1)) < eps {
		return false
	}
	//fmt.Printf("\rFactorization matrix: 100%%\n")
	//fmt.Printf("\rSolving a system of equations: ")
	b.SetVec(size-1, b.AtVec(size-1)/a.At(size-1, size-1))
	for i := size - 2; i >= 0; i-- {
		//if i%10 == 0 {
		//	fmt.Printf("\rSolving a system of equations:: %d%%", 100*(size-2-i)/(size-1))
		//}
		for j := i + 1; j < size; j++ {
			b.SetVec(i, b.AtVec(i)-b.AtVec(j)*a.At(i, j))
		}
		if math.Abs(a.At(i, i)) < eps {
			return false
		}
		b.SetVec(i, b.AtVec(i)/a.At(i, i))
	}
	// fmt.Printf("\rSolving a system of equations: 100%%\nDone\n")
	return true
}

func CreateShape(size int, shape func(int, int) float64) (*mat.Dense, error) {
	a := mat.NewDense(size, size, nil)
	b := mat.NewVecDense(size, nil)
	c := mat.NewDense(size, size, nil)

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			for k := 0; k < size; k++ {
				a.Set(j, k, shape(j, k))
			}
			if i == j {
				b.SetVec(j, 1.0)
			} else {
				b.SetVec(j, 0.0)
			}
		}
		if !Gauss(a, b, 1.0e-10) {
			return nil, fmt.Errorf("bad finite element")
		}
		for j := 0; j < size; j++ {
			c.Set(j, i, b.AtVec(j))
		}
	}
	return c, nil
}

// Volume1d2 - segment length
func Volume1d2(x *mat.Dense) float64 {
	_, cols := x.Dims()
	v := 0.0
	for j := 0; j < cols; j++ {
		v += math.Pow(x.At(0, j)-x.At(1, j), 2)
	}
	return math.Sqrt(v)
}

// Volume2d3 - area of triangle 2d
func Volume2d3(x *mat.Dense) float64 {
	_, cols := x.Dims()
	a, b, c := 0.0, 0.0, 0.0
	for j := 0; j < cols; j++ {
		a += math.Pow(x.At(0, j)-x.At(1, j), 2)
		b += math.Pow(x.At(0, j)-x.At(2, j), 2)
		c += math.Pow(x.At(2, j)-x.At(1, j), 2)
	}
	a, b, c = math.Sqrt(a), math.Sqrt(b), math.Sqrt(c)
	p := 0.5 * (a + b + c)
	return math.Sqrt(p * (p - a) * (p - b) * (p - c))
}

// Volume2d4 - area of quadrilateral
func Volume2d4(x *mat.Dense) float64 {
	_, cols := x.Dims()
	a, b, c, d, e, f := 0.0, 0.0, 0.0, 0.0, 0.0, 0.0
	for j := 0; j < cols; j++ {
		a += math.Pow(x.At(0, j)-x.At(1, j), 2)
		b += math.Pow(x.At(1, j)-x.At(2, j), 2)
		c += math.Pow(x.At(2, j)-x.At(3, j), 2)
		d += math.Pow(x.At(3, j)-x.At(0, j), 2)
		e += math.Pow(x.At(0, j)-x.At(2, j), 2)
		f += math.Pow(x.At(1, j)-x.At(3, j), 2)
	}
	a, b, c, d, e, f = math.Sqrt(a), math.Sqrt(b), math.Sqrt(c), math.Sqrt(d), math.Sqrt(e), math.Sqrt(f)
	p := 0.5 * (a + b + c + d)
	return math.Sqrt((p-a)*(p-b)*(p-c)*(p-d) + 0.25*(e*f+a*c+b*d)*(e*f-a*c-b*d))
}

// Volume3d4 - pyramid volume
func Volume3d4(x *mat.Dense) float64 {
	mt := mat.NewDense(3, 3, []float64{
		x.At(1, 0) - x.At(0, 0), x.At(1, 1) - x.At(0, 1), x.At(1, 2) - x.At(0, 2),
		x.At(2, 0) - x.At(0, 0), x.At(2, 1) - x.At(0, 1), x.At(2, 2) - x.At(0, 2),
		x.At(3, 0) - x.At(0, 0), x.At(3, 1) - x.At(0, 1), x.At(3, 2) - x.At(0, 2),
	})
	return math.Abs(mat.Det(mt)) / 6.0
}

// Volume3d8 - volume of a quadrangular hexagon
func Volume3d8(x *mat.Dense) float64 {
	ref := [6][4]int{{0, 1, 4, 7}, {4, 1, 5, 7}, {1, 2, 6, 7}, {1, 5, 6, 7}, {1, 2, 3, 7}, {0, 3, 1, 7}}
	v := 0.0
	for i := 0; i < 6; i++ {
		mt := mat.NewDense(3, 3, []float64{
			x.At(ref[i][1], 0) - x.At(ref[i][0], 0), x.At(ref[i][1], 1) - x.At(ref[i][0], 1), x.At(ref[i][1], 2) - x.At(ref[i][0], 2),
			x.At(ref[i][2], 0) - x.At(ref[i][0], 0), x.At(ref[i][2], 1) - x.At(ref[i][0], 1), x.At(ref[i][2], 2) - x.At(ref[i][0], 2),
			x.At(ref[i][3], 0) - x.At(ref[i][0], 0), x.At(ref[i][3], 1) - x.At(ref[i][0], 1), x.At(ref[i][3], 2) - x.At(ref[i][0], 2),
		})
		v += math.Abs(mat.Det(mt)) / 6.0
	}
	return v
}

// TransformMatrix - create transformation matrix for shell finite element
func TransformMatrix(x *mat.Dense) *mat.Dense {
	m := mat.NewDense(3, 3, nil)
	tmp := createVector3(mat.VecDenseCopyOf(x.RowView(2)), mat.VecDenseCopyOf(x.RowView(0)))
	vx := createVector3(mat.VecDenseCopyOf(x.RowView(1)), mat.VecDenseCopyOf(x.RowView(0)))
	vz := crossProduct3(vx, tmp)
	vy := crossProduct3(vz, vx)
	for i := 0; i < 3; i++ {
		m.Set(0, i, vx.AtVec(i))
		m.Set(1, i, vy.AtVec(i))
		m.Set(2, i, vz.AtVec(i))
	}
	return m
}

func ExtTransformMatrix(m *mat.Dense, size int) *mat.Dense {
	res := mat.NewDense(size, size, nil)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < size; k += 3 {
				res.Set(i+k, j+k, m.At(i, j))
			}
		}
	}
	return res
}

func createVector3(xi, xj *mat.VecDense) *mat.VecDense {
	res := mat.NewVecDense(3, nil)
	for k := 0; k < 3; k++ {
		res.SetVec(k, xj.AtVec(k)-xi.AtVec(k))
	}
	return norm3(res)
}

func crossProduct3(a, b *mat.VecDense) *mat.VecDense {
	res := mat.NewVecDense(3, nil)
	res.SetVec(0, a.AtVec(1)*b.AtVec(2)-a.AtVec(2)*b.AtVec(1))
	res.SetVec(1, a.AtVec(2)*b.AtVec(0)-a.AtVec(0)*b.AtVec(2))
	res.SetVec(2, a.AtVec(0)*b.AtVec(1)-a.AtVec(1)*b.AtVec(0))
	return norm3(res)
}

func norm3(v *mat.VecDense) *mat.VecDense {
	norm := math.Sqrt(v.AtVec(0)*v.AtVec(0) + v.AtVec(1)*v.AtVec(1) + v.AtVec(2)*v.AtVec(2))
	for k := 0; k < 3; k++ {
		v.SetVec(k, v.AtVec(k)/norm)
	}
	return v
}

func Mul(lhs, rhs mat.Matrix) *mat.Dense {
	rows, _ := lhs.Dims()
	_, cols := rhs.Dims()
	res := mat.NewDense(rows, cols, nil)
	res.Mul(lhs, rhs)
	return res
}

func Add(lhs, rhs mat.Matrix) *mat.Dense {
	rows, cols := lhs.Dims()
	res := mat.NewDense(rows, cols, nil)
	res.Add(lhs, rhs)
	return res
}

func Scale(lhs float64, rhs mat.Matrix) *mat.Dense {
	rows, cols := rhs.Dims()
	res := mat.NewDense(rows, cols, nil)
	res.Scale(lhs, rhs)
	return res
}

func Transpose(m *mat.Dense) *mat.Dense {
	rows, cols := m.Dims()
	res := mat.NewDense(cols, rows, nil)
	res.Copy(m.T())
	return res
}
