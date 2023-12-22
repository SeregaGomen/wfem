package fe

import (
	"gonum.org/v1/gonum/mat"
	"math"
	"wfem/cmd/fem/util"
)

type FiniteElementParameters struct {
	YoungModulus  float64
	PoissonRation float64
	Thickness     float64
}

type FiniteElement interface {
	Calculate(*mat.VecDense) *mat.Dense
	Create() *mat.Dense
}

type FE struct {
	size    int
	freedom int
	FiniteElementParameters
}

type FiniteElement1D struct {
	FE
	ShapeFunction1D
}

func NewFE1D(shape ShapeFunction1D, params FiniteElementParameters) *FiniteElement1D {
	return &FiniteElement1D{FE: FE{size: shape.Size(), freedom: 1, FiniteElementParameters: params}, ShapeFunction1D: shape}
}

func (f *FiniteElement1D) Create() *mat.Dense {
	res := mat.NewDense(f.size, f.size, nil)
	// Jacobian and inverted Jacobi matrix
	jacobian := mat.Det(mat.NewDense(1, 1, []float64{(f.X().At(1, 0) - f.X().At(0, 0)) * 0.5}))
	invJacobi := 1.0 / jacobian
	// Numerical integration according to the Gauss formula on the interval [-0.5; 0.5]
	for i := 0; i < len(*f.W()); i++ {
		// Gradient Matrix
		B := mat.NewDense(1, f.size, nil)
		for j := 0; j < f.size; j++ {
			B.Set(0, f.freedom*j, invJacobi*f.ShapeDxi(i, j))
		}
		// Calculation of the local stiffness matrix
		K := util.Scale((*f.W())[i]*f.Thickness*math.Abs(jacobian), util.Mul(util.Scale(f.YoungModulus, B.T()), B))
		res.Add(res, K)
	}
	return res
}

func (f *FiniteElement1D) Calculate(u *mat.VecDense) *mat.Dense {
	res := mat.NewDense(2, 2, nil)
	B := mat.NewDense(1, f.size, nil)
	for i := 0; i < f.size; i++ {
		for j := 0; j < f.size; j++ {
			B.Set(0, f.freedom*j, f.ShapeDx(i, j))
			strain := mat.NewDense(1, 1, nil)
			stress := mat.NewDense(1, 1, nil)
			strain.Mul(B, u)
			stress.Scale(f.YoungModulus, strain)
			res.Set(0, i, res.At(0, i)+strain.At(0, 0))
			res.Set(1, i, res.At(1, i)+strain.At(0, 0))
		}
	}
	return res
}

type FiniteElement2D struct {
	FE
	ShapeFunction2D
}

func NewFE2D(shape ShapeFunction2D, params FiniteElementParameters) *FiniteElement2D {
	return &FiniteElement2D{FE: FE{size: shape.Size(), freedom: 2, FiniteElementParameters: params}, ShapeFunction2D: shape}
}

func (f *FiniteElement2D) Create() *mat.Dense {
	var invJacobi mat.Dense
	res := mat.NewDense(f.size*f.freedom, f.size*f.freedom, nil)
	// Numerical integration according to the Gauss formula on the interval [-0.5; 0.5]
	for i := 0; i < len(*f.W()); i++ {
		jacobi := mat.NewDense(2, 2, nil)
		for j := 0; j < 2; j++ {
			for k := 0; k < f.size; k++ {
				jacobi.Set(0, j, jacobi.At(0, j)+f.ShapeDxi(i, k)*f.X().At(k, j))
				jacobi.Set(1, j, jacobi.At(1, j)+f.ShapeDeta(i, k)*f.X().At(k, j))
			}
		}
		// Jacobian and inverted Jacobi matrix
		jacobian := mat.Det(jacobi)
		_ = invJacobi.Inverse(jacobi)
		// Gradient Matrix
		B := mat.NewDense(3, f.size*f.freedom, nil)
		for j := 0; j < f.size; j++ {
			B.Set(0, f.freedom*j+0, invJacobi.At(0, 0)*f.ShapeDxi(i, j)+invJacobi.At(0, 1)*f.ShapeDeta(i, j))
			B.Set(2, f.freedom*j+1, B.At(0, f.freedom*j+0))
			B.Set(1, f.freedom*j+1, invJacobi.At(1, 0)*f.ShapeDxi(i, j)+invJacobi.At(1, 1)*f.ShapeDeta(i, j))
			B.Set(2, f.freedom*j+0, B.At(1, f.freedom*j+1))
		}
		// Calculation of the local stiffness matrix
		K := util.Scale((*f.W())[i]*f.Thickness*math.Abs(jacobian), util.Mul(util.Mul(B.T(), f.elasticMatrix()), B))
		res.Add(res, K)
	}
	return res
}

func (f *FiniteElement2D) elasticMatrix() *mat.Dense {
	return mat.NewDense(3, 3, []float64{
		f.YoungModulus / (1.0 - f.PoissonRation*f.PoissonRation), f.PoissonRation * f.YoungModulus / (1.0 - f.PoissonRation*f.PoissonRation), 0.0,
		f.PoissonRation * f.YoungModulus / (1.0 - f.PoissonRation*f.PoissonRation), f.YoungModulus / (1.0 - f.PoissonRation*f.PoissonRation), 0.0,
		0.0, 0.0, 0.5 * (1.0 - f.PoissonRation) * f.YoungModulus / (1.0 - f.PoissonRation*f.PoissonRation),
	})
}

func (f *FiniteElement2D) Calculate(u *mat.VecDense) *mat.Dense {
	res := mat.NewDense(8, f.size, nil)
	B := mat.NewDense(3, f.size*f.freedom, nil)
	var stress mat.Dense
	var strain mat.Dense
	for i := 0; i < f.size; i++ {
		for j := 0; j < f.size; j++ {
			B.Set(0, f.freedom*j+0, f.ShapeDx(i, j))
			B.Set(2, f.freedom*j+1, B.At(0, f.freedom*j+0))
			B.Set(1, f.freedom*j+1, f.ShapeDy(i, j))
			B.Set(2, f.freedom*j+0, B.At(1, f.freedom*j+1))
		}
		strain.Mul(B, u)
		stress.Mul(f.elasticMatrix(), &strain)
		for j := 0; j < 3; j++ {
			res.Set(j, i, res.At(j, i)+strain.At(j, 0))
			res.Set(j+3, i, res.At(j+3, i)+stress.At(j, 0))
		}
	}
	return res
}

type FiniteElement3D struct {
	FE
	ShapeFunction3D
}

func NewFE3D(shape ShapeFunction3D, params FiniteElementParameters) *FiniteElement3D {
	return &FiniteElement3D{FE: FE{size: shape.Size(), freedom: 3, FiniteElementParameters: params}, ShapeFunction3D: shape}
}

func (f *FiniteElement3D) elasticMatrix() *mat.Dense {
	e := f.YoungModulus
	m := f.PoissonRation
	return mat.NewDense(6, 6, []float64{
		e * (1.0 - m) / (1.0 + m) / (1.0 - 2.0*m), m / (1.0 - m) * e * (1.0 - m) / (1.0 + m) / (1.0 - 2.0*m), m / (1.0 - m) * e * (1.0 - m) / (1.0 + m) / (1.0 - 2.0*m), 0.0, 0.0, 0.0,
		m / (1.0 - m) * e * (1.0 - m) / (1.0 + m) / (1.0 - 2.0*m), e * (1.0 - m) / (1.0 + m) / (1.0 - 2.0*m), m / (1.0 - m) * e * (1.0 - m) / (1.0 + m) / (1.0 - 2.0*m), 0.0, 0.0, 0.0,
		m / (1.0 - m) * e * (1.0 - m) / (1.0 + m) / (1.0 - 2.0*m), m / (1.0 - m) * e * (1.0 - m) / (1.0 + m) / (1.0 - 2.0*m), e * (1.0 - m) / (1.0 + m) / (1.0 - 2.0*m), 0.0, 0.0, 0.0,
		0.0, 0.0, 0.0, 0.5 * (1.0 - 2.0*m) / (1.0 - m) * e * (1.0 - m) / (1.0 + m) / (1.0 - 2.0*m), 0.0, 0.0,
		0.0, 0.0, 0.0, 0.0, 0.5 * (1.0 - 2.0*m) / (1.0 - m) * e * (1.0 - m) / (1.0 + m) / (1.0 - 2.0*m), 0.0,
		0.0, 0.0, 0.0, 0.0, 0.0, 0.5 * (1.0 - 2.0*m) / (1.0 - m) * e * (1.0 - m) / (1.0 + m) / (1.0 - 2.0*m),
	})
}

func (f *FiniteElement3D) Create() *mat.Dense {
	var invJacobi mat.Dense
	res := mat.NewDense(f.size*f.freedom, f.size*f.freedom, nil)
	// Numerical integration according to the Gauss formula
	for i := 0; i < len(*f.W()); i++ {
		jacobi := mat.NewDense(3, 3, nil)
		for j := 0; j < 3; j++ {
			for k := 0; k < f.size; k++ {
				jacobi.Set(0, j, jacobi.At(0, j)+f.ShapeDxi(i, k)*f.X().At(k, j))
				jacobi.Set(1, j, jacobi.At(1, j)+f.ShapeDeta(i, k)*f.X().At(k, j))
				jacobi.Set(2, j, jacobi.At(2, j)+f.ShapeDpsi(i, k)*f.X().At(k, j))
			}
		}
		// Jacobian and inverted Jacobi matrix
		jacobian := mat.Det(jacobi)
		_ = invJacobi.Inverse(jacobi)
		// Gradient Matrix
		B := mat.NewDense(6, f.size*f.freedom, nil)
		for j := 0; j < f.size; j++ {
			B.Set(0, f.freedom*j+0, invJacobi.At(0, 0)*f.ShapeDxi(i, j)+invJacobi.At(0, 1)*f.ShapeDeta(i, j)+invJacobi.At(0, 2)*f.ShapeDpsi(i, j))
			B.Set(3, f.freedom*j+1, B.At(0, f.freedom*j+0))
			B.Set(5, f.freedom*j+2, B.At(0, f.freedom*j+0))
			B.Set(1, f.freedom*j+1, invJacobi.At(1, 0)*f.ShapeDxi(i, j)+invJacobi.At(1, 1)*f.ShapeDeta(i, j)+invJacobi.At(1, 2)*f.ShapeDpsi(i, j))
			B.Set(3, f.freedom*j+0, B.At(1, f.freedom*j+1))
			B.Set(4, f.freedom*j+2, B.At(1, f.freedom*j+1))
			B.Set(2, f.freedom*j+2, invJacobi.At(2, 0)*f.ShapeDxi(i, j)+invJacobi.At(2, 1)*f.ShapeDeta(i, j)+invJacobi.At(2, 2)*f.ShapeDpsi(i, j))
			B.Set(4, f.freedom*j+1, B.At(2, f.freedom*j+2))
			B.Set(5, f.freedom*j+0, B.At(2, f.freedom*j+2))
		}
		// Calculation of the local stiffness matrix
		K := util.Scale((*f.W())[i]*math.Abs(jacobian), util.Mul(util.Mul(B.T(), f.elasticMatrix()), B))
		res.Add(res, K)
	}
	return res
}

func (f *FiniteElement3D) Calculate(u *mat.VecDense) *mat.Dense {
	res := mat.NewDense(15, f.size, nil)
	B := mat.NewDense(6, f.size*f.freedom, nil)
	var stress mat.Dense
	var strain mat.Dense
	for i := 0; i < f.size; i++ {
		for j := 0; j < f.size; j++ {
			B.Set(0, f.freedom*j+0, f.ShapeDx(i, j))
			B.Set(3, f.freedom*j+1, B.At(0, f.freedom*j+0))
			B.Set(5, f.freedom*j+2, B.At(0, f.freedom*j+0))
			B.Set(1, f.freedom*j+1, f.ShapeDy(i, j))
			B.Set(3, f.freedom*j+0, B.At(1, f.freedom*j+1))
			B.Set(4, f.freedom*j+2, B.At(1, f.freedom*j+1))
			B.Set(2, f.freedom*j+2, f.ShapeDz(i, j))
			B.Set(4, f.freedom*j+1, B.At(2, f.freedom*j+2))
			B.Set(5, f.freedom*j+0, B.At(2, f.freedom*j+2))
		}
		strain.Mul(B, u)
		stress.Mul(f.elasticMatrix(), &strain)
		for j := 0; j < 6; j++ {
			res.Set(j, i, res.At(j, i)+strain.At(j, 0))
			res.Set(j+6, i, res.At(j+6, i)+stress.At(j, 0))
		}
	}
	return res
}

type FiniteElement3DS struct {
	transformMatrix *mat.Dense
	FiniteElement2D
	ShapeFunction2D
}

func NewFE3DS(shape ShapeFunction2D, transformMatrix *mat.Dense, params FiniteElementParameters) *FiniteElement3DS {
	return &FiniteElement3DS{
		transformMatrix: transformMatrix,
		FiniteElement2D: FiniteElement2D{FE: FE{size: shape.Size(), freedom: 6, FiniteElementParameters: params}}, ShapeFunction2D: shape,
	}
}

func (f *FiniteElement3DS) extraElasticMatrix() *mat.Dense {
	return mat.NewDense(2, 2, []float64{f.YoungModulus / (2.0 + 2.0*f.PoissonRation), 0.0, 0.0, f.YoungModulus / (2.0 + 2.0*f.PoissonRation)})
}

func (f *FiniteElement3DS) Create() *mat.Dense {
	var invJacobi mat.Dense
	res := mat.NewDense(f.size*f.freedom, f.size*f.freedom, nil)
	// Numerical integration according to the Gauss formula
	for i := 0; i < len(*f.W()); i++ {
		// Jacobian and inverted Jacobi matrix
		jacobi := mat.NewDense(2, 2, nil)
		for j := 0; j < 2; j++ {
			for k := 0; k < f.size; k++ {
				jacobi.Set(0, j, jacobi.At(0, j)+f.ShapeDxi(i, k)*f.X().At(k, j))
				jacobi.Set(1, j, jacobi.At(1, j)+f.ShapeDeta(i, k)*f.X().At(k, j))
			}
		}
		jacobian := mat.Det(jacobi)
		_ = invJacobi.Inverse(jacobi)
		// Gradient Matrix
		bm := mat.NewDense(3, f.size*f.freedom, nil)
		bp := mat.NewDense(3, f.size*f.freedom, nil)
		bc := mat.NewDense(2, f.size*f.freedom, nil)
		for j := 0; j < f.size; j++ {
			bm.Set(0, f.freedom*j+0, invJacobi.At(0, 0)*f.ShapeDxi(i, j)+invJacobi.At(0, 1)*f.ShapeDeta(i, j))
			bm.Set(2, f.freedom*j+1, bm.At(0, f.freedom*j+0))
			bp.Set(0, f.freedom*j+3, bm.At(0, f.freedom*j+0))
			bp.Set(2, f.freedom*j+4, bm.At(0, f.freedom*j+0))
			bc.Set(0, f.freedom*j+2, bm.At(0, f.freedom*j+0))
			bm.Set(1, f.freedom*j+1, invJacobi.At(1, 0)*f.ShapeDxi(i, j)+invJacobi.At(1, 1)*f.ShapeDeta(i, j))
			bm.Set(2, f.freedom*j+0, bm.At(1, f.freedom*j+1))
			bp.Set(1, f.freedom*j+4, bm.At(1, f.freedom*j+1))
			bp.Set(2, f.freedom*j+3, bm.At(1, f.freedom*j+1))
			bc.Set(1, f.freedom*j+2, bm.At(1, f.freedom*j+1))
			bc.Set(0, f.freedom*j+3, f.Shape(i, j))
			bc.Set(1, f.freedom*j+4, bc.At(0, f.freedom*j+3))
		}
		// Calculation of the local stiffness matrix
		K := util.Add(util.Add(util.Scale(f.Thickness, util.Mul(util.Mul(bm.T(), f.elasticMatrix()), bm)),
			util.Scale(math.Pow(f.Thickness, 3)/12.0, util.Mul(util.Mul(bp.T(), f.elasticMatrix()), bp))),
			util.Scale(f.Thickness*5.0/6.0, util.Mul(util.Mul(bc.T(), f.extraElasticMatrix()), bc)))
		res.Add(res, util.Scale((*f.W())[i]*math.Abs(jacobian), K))
	}
	// Finding the maximum diagonal element
	singular := 0.0
	for i := 0; i < f.size*f.freedom; i++ {
		if res.At(i, i) > singular {
			singular = res.At(i, i)
		}
	}
	singular *= 1.0e-3
	// Removing the singularity
	for i := 0; i < f.size; i++ {
		res.Set(f.freedom*(i+1)-1, f.freedom*(i+1)-1, singular)
	}
	// Convert from local to global coordinates
	m := util.ExtTransformMatrix(f.transformMatrix, f.size*f.freedom)
	res = util.Mul(util.Mul(m.T(), res), m)
	return res
}

func (f *FiniteElement3DS) Calculate(u *mat.VecDense) *mat.Dense {
	res := mat.NewDense(18, f.size, nil)
	lu := util.Mul(util.ExtTransformMatrix(f.transformMatrix, f.size*f.freedom), u)
	index := [6][2]int{{0, 0}, {1, 1}, {2, 2}, {0, 1}, {0, 2}, {1, 2}}
	for i := 0; i < f.size; i++ {
		bm := mat.NewDense(3, f.size*f.freedom, nil)
		bp := mat.NewDense(3, f.size*f.freedom, nil)
		bc := mat.NewDense(2, f.size*f.freedom, nil)
		for j := 0; j < f.size; j++ {
			bm.Set(0, f.freedom*j+0, f.ShapeDx(i, j))
			bm.Set(2, f.freedom*j+1, bm.At(0, f.freedom*j+0))
			bp.Set(0, f.freedom*j+3, bm.At(0, f.freedom*j+0))
			bp.Set(2, f.freedom*j+4, bm.At(0, f.freedom*j+0))
			bc.Set(0, f.freedom*j+2, bm.At(0, f.freedom*j+0))
			bm.Set(1, f.freedom*j+1, f.ShapeDy(i, j))
			bm.Set(2, f.freedom*j+0, bm.At(1, f.freedom*j+1))
			bp.Set(1, f.freedom*j+4, bm.At(1, f.freedom*j+1))
			bp.Set(2, f.freedom*j+3, bm.At(1, f.freedom*j+1))
			bc.Set(1, f.freedom*j+2, bm.At(1, f.freedom*j+1))
			if i == j {
				bc.Set(0, f.freedom*j+3, 1.0)
				bc.Set(1, f.freedom*j+4, 1.0)
			}
		}
		strainM := util.Mul(bm, lu)
		strainP := util.Mul(bp, lu)
		strainC := util.Mul(bc, lu)
		stressM := util.Mul(f.elasticMatrix(), strainM)
		stressP := util.Scale(f.Thickness*0.5, util.Mul(f.elasticMatrix(), strainP))
		stressC := util.Mul(f.extraElasticMatrix(), strainC)

		localStrain := mat.NewDense(3, 3, []float64{
			strainM.At(0, 0) + strainP.At(0, 0), strainM.At(2, 0) + strainP.At(2, 0), strainC.At(0, 0),
			strainM.At(2, 0) + strainP.At(2, 0), strainM.At(1, 0) + strainP.At(1, 0), strainC.At(1, 0),
			strainC.At(0, 0), strainC.At(1, 0), 0.0,
		})
		localStress := mat.NewDense(3, 3, []float64{
			stressM.At(0, 0) + stressP.At(0, 0), stressM.At(2, 0) + stressP.At(2, 0), stressC.At(0, 0),
			stressM.At(2, 0) + stressP.At(2, 0), stressM.At(1, 0) + stressP.At(1, 0), stressC.At(1, 0),
			stressC.At(0, 0), stressC.At(1, 0), 0.0,
		})
		globalStrain := util.Mul(util.Mul(f.transformMatrix.T(), localStrain), f.transformMatrix)
		globalStress := util.Mul(util.Mul(f.transformMatrix.T(), localStress), f.transformMatrix)
		for j := 0; j < 6; j++ {
			res.Set(j, i, globalStrain.At(index[j][0], index[j][1]))
			res.Set(j+6, i, globalStress.At(index[j][0], index[j][1]))
		}
	}
	return res
}
