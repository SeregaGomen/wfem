package fem

import (
	"fmt"
	"math"
	"os"
	"sync/atomic"
	"time"
	"wfem/cmd/fem/fe"
	"wfem/cmd/fem/mesh"
	"wfem/cmd/fem/params"
	"wfem/cmd/fem/progress"
	"wfem/cmd/fem/solver"
	"wfem/cmd/fem/util"

	"gonum.org/v1/gonum/mat"
)

type MatrixData struct {
	index  int
	matrix *mat.Dense
}

type VectorData struct {
	index  int
	direct int
	vector [3]float64
}

type StaticFEM struct {
	res    *mat.Dense
	solver solver.Solver
	mesh   mesh.Mesh
	params params.FEMParameters
}

func NewStaticFEM() StaticFEM {
	return StaticFEM{params: params.New()}
}

func (fem *StaticFEM) SetMesh(name string) error {
	return fem.mesh.Load(name)
}

func (fem *StaticFEM) AddBoundaryCondition(value, predicate string, direct int) {
	fem.params.AddBoundaryCondition(value, predicate, direct)
}

func (fem *StaticFEM) AddVolumeLoad(value, predicate string, direct int) {
	fem.params.AddVolumeLoad(value, predicate, direct)
}

func (fem *StaticFEM) AddSurfaceLoad(value, predicate string, direct int) {
	fem.params.AddSurfaceLoad(value, predicate, direct)
}

func (fem *StaticFEM) AddPointLoad(value, predicate string, direct int) {
	fem.params.AddConcentratedLoad(value, predicate, direct)
}

func (fem *StaticFEM) AddPressureLoad(value, predicate string) {
	fem.params.AddPressureLoad(value, predicate)
}

func (fem *StaticFEM) AddThickness(value, predicate string) {
	fem.params.AddThickness(value, predicate)
}

func (fem *StaticFEM) AddPoissonRatio(value, predicate string) {
	fem.params.AddPoissonRatio(value, predicate)
}

func (fem *StaticFEM) AddYoungModulus(value, predicate string) {
	fem.params.AddYoungModulus(value, predicate)
}

func (fem *StaticFEM) AddVariable(name string, value float64) {
	fem.params.AddVariable(name, value)
}

func (fem *StaticFEM) SetEps(eps float64) {
	fem.params.SetEps(eps)
}

func (fem *StaticFEM) SetNumThread(num int) {
	fem.params.SetNumThread(num)
}

func (fem *StaticFEM) Calculate() error {
	var err error
	fmt.Printf("Using threads: %d\n", fem.params.NumThread)
	start := time.Now()
	fem.solver = solver.NewDenseSolver(&fem.mesh)
	//fem.solver = solver.NewEigenSolver(&fem.mesh)
	if err = fem.calcGlobalMatrix(); err != nil {
		return err
	}
	if err = fem.addPointLoad(); err != nil {
		return err
	}
	if err = fem.addVolumeLoad(); err != nil {
		return err
	}
	if err = fem.addSurfaceLoad(); err != nil {
		return err
	}
	if err = fem.addBoundaryCondition(); err != nil {
		return err
	}
	x, err := fem.solver.Solve()
	if err != nil {
		return err
	}
	err = fem.calcResult(x)
	if err != nil {
		return err
	}
	fem.printSummary()
	duration := time.Since(start)
	fmt.Printf("Lead time: %0.2f sec\n\n", duration.Seconds())
	return nil
}

func (fem *StaticFEM) calcGlobalMatrix() error {
	data := make(chan MatrixData, fem.params.NumThread)
	errChan := make(chan error, fem.params.NumThread)
	done := make(chan struct{})
	step := fem.mesh.NumFE() / fem.params.NumThread
	go func() {
		var local MatrixData
		freedom := fem.mesh.Freedom()
		size := fem.mesh.FeSize() * freedom
		msg := progress.NewProgress("Building a global stiffness matrix", 0, fem.mesh.NumFE(), 10)
		for k := 0; k < fem.mesh.NumFE(); k++ {
			msg.AddProgress()
			local = <-data
			for i := 0; i < size; i++ {
				for j := i; j < size; j++ {
					fem.solver.AddMatrix(fem.mesh.FE[local.index][i/freedom]*freedom+i%freedom, fem.mesh.FE[local.index][j/freedom]*freedom+j%freedom, local.matrix.At(i, j))
					if i != j {
						fem.solver.AddMatrix(fem.mesh.FE[local.index][j/freedom]*freedom+j%freedom, fem.mesh.FE[local.index][i/freedom]*freedom+i%freedom, local.matrix.At(i, j))
					}
				}
			}
		}
		done <- struct{}{}
	}()
	for i := 0; i < fem.params.NumThread; i++ {
		begin := i * step
		end := (i + 1) * step
		if i == fem.params.NumThread-1 {
			end = fem.mesh.NumFE()
		}
		go func() {
			var elm fe.FiniteElement
			var err error
			defer func() {
				errChan <- err
			}()
			for j := begin; j < end; j++ {
				if elm, err = fem.createFE(j); err != nil {
					return
				}
				local := elm.Create()
				data <- MatrixData{index: j, matrix: local}
			}
		}()
	}
	if err := <-errChan; err != nil {
		return err
	}
	<-done
	return nil
}

func (fem *StaticFEM) GetResult() *mat.Dense {
	return fem.res
}

func (fem *StaticFEM) printSummary() {
	fmt.Println("----------------------------------------------")
	fmt.Println("Fun:\tmin\t\tmax")
	for i, name := range *fem.ResultNames() {
		fmt.Printf("%s\t%+e\t%+e\n", name, mat.Min(fem.res.RowView(i)), mat.Max(fem.res.RowView(i)))
	}
}

func (fem *StaticFEM) addBoundaryCondition() error {
	if !fem.params.FindParameter(params.BoundaryCondition) {
		return nil
	}
	var counter int32
	done := make(chan struct{})
	data := make(chan VectorData, fem.params.NumThread)
	errChan := make(chan error, fem.params.NumThread)
	step := fem.mesh.NumVertex() / fem.params.NumThread
	go fem.addData(data, done, fem.solver.SetBoundaryCondition)
	msg := progress.NewProgress("Using of boundary conditions", 0, fem.mesh.NumVertex(), 10)
	for i := 0; i < fem.params.NumThread; i++ {
		begin := i * step
		end := (i + 1) * step
		if i == fem.params.NumThread-1 {
			end = fem.mesh.NumVertex()
		}
		go func() {
			var err error
			var ok bool
			var value float64
			defer func() {
				errChan <- err
			}()
			for j := begin; j < end; j++ {
				msg.AddProgress()
				for k := range fem.params.Params {
					if fem.params.Params[k].Type == params.BoundaryCondition {
						x := mat.NewVecDense(fem.mesh.FeDim(), fem.mesh.X[j])
						if len(fem.params.Params[k].Predicate) > 0 {
							ok, err = fem.params.Params[k].GetPredicate(x, &fem.params.Variables)
							if err != nil {
								errChan <- err
								return
							}
							if !ok {
								continue
							}
						}
						value, err = fem.params.Params[k].GetValue(x, &fem.params.Variables)
						if err != nil {
							errChan <- err
							return
						}
						data <- VectorData{index: j, direct: fem.params.Params[k].Direct, vector: [3]float64{value, value, value}}
					}
				}
				atomic.AddInt32(&counter, 1)
				if int(counter) == fem.mesh.NumVertex() {
					close(data)
				}
			}
		}()
	}
	if err := <-errChan; err != nil {
		return err
	}
	<-done
	return nil
}

func (fem *StaticFEM) addPointLoad() error {
	if !fem.params.FindParameter(params.PointLoad) {
		return nil
	}
	var counter int32
	done := make(chan struct{})
	data := make(chan VectorData, fem.params.NumThread)
	errChan := make(chan error, fem.params.NumThread)
	step := fem.mesh.NumVertex() / fem.params.NumThread
	go fem.addData(data, done, fem.solver.AddVector)
	msg := progress.NewProgress("Calculation of point loads", 0, fem.mesh.NumVertex(), 10)
	for i := 0; i < fem.params.NumThread; i++ {
		begin := i * step
		end := (i + 1) * step
		if i == fem.params.NumThread-1 {
			end = fem.mesh.NumVertex()
		}
		go func() {
			var err error
			var ok bool
			var value float64
			defer func() {
				errChan <- err
			}()
			for j := begin; j < end; j++ {
				msg.AddProgress()
				for k := range fem.params.Params {
					if fem.params.Params[k].Type == params.PointLoad {
						x := mat.NewVecDense(fem.mesh.FeDim(), fem.mesh.X[j])
						if len(fem.params.Params[k].Predicate) > 0 {
							ok, err = fem.params.Params[k].GetPredicate(x, &fem.params.Variables)
							if err != nil {
								errChan <- err
								return
							}
							if !ok {
								continue
							}
						}
						value, err = fem.params.Params[k].GetValue(x, &fem.params.Variables)
						if err != nil {
							errChan <- err
							return
						}
						data <- VectorData{index: j, direct: fem.params.Params[k].Direct, vector: [3]float64{value, value, value}}
					}
				}
				atomic.AddInt32(&counter, 1)
				if int(counter) == fem.mesh.NumVertex() {
					close(data)
				}
			}
		}()
	}
	if err := <-errChan; err != nil {
		return err
	}
	<-done
	return nil
}

func (fem *StaticFEM) addData(data chan VectorData, done chan struct{}, fun func(int, float64)) {
	var chData VectorData
	var ok bool
	direct := [3]int{params.X, params.Y, params.Z}
	for {
		if chData, ok = <-data; !ok {
			done <- struct{}{}
			break
		}
		for l := 0; l < fem.mesh.FeDim(); l++ {
			if chData.direct&direct[l] == direct[l] {
				fun(chData.index*fem.mesh.Freedom()+l, chData.vector[l])
			}
		}
	}
}

func (fem *StaticFEM) addVolumeLoad() error {
	if !fem.params.FindParameter(params.VolumeLoad) {
		return nil
	}
	var share []float64
	switch fem.mesh.FeType {
	case mesh.Fe1d2:
		share = []float64{0.5, 0.5}
	case mesh.Fe2d3:
		fallthrough
	case mesh.Fe3d3s:
		share = []float64{0.33333333333, 0.33333333333, 0.33333333333}
	case mesh.Fe2d4:
		fallthrough
	case mesh.Fe3d4:
		fallthrough
	case mesh.Fe3d4s:
		share = []float64{0.25, 0.25, 0.25, 0.25}
	case mesh.Fe3d8:
		share = []float64{0.125, 0.125, 0.125, 0.125, 0.125, 0.125, 0.125, 0.125}
	}
	var counter int32
	done := make(chan struct{})
	data := make(chan VectorData, fem.params.NumThread)
	errChan := make(chan error, fem.params.NumThread)
	step := fem.mesh.NumFE() / fem.params.NumThread
	go fem.addData(data, done, fem.solver.AddVector)
	msg := progress.NewProgress("Calculation of volume loads", 0, fem.mesh.NumFE(), 10)
	for i := 0; i < fem.params.NumThread; i++ {
		begin := i * step
		end := (i + 1) * step
		if i == fem.params.NumThread-1 {
			end = fem.mesh.NumFE()
		}
		go func() {
			var err error
			var ok bool
			var value float64
			defer func() {
				errChan <- err
			}()
			for j := begin; j < end; j++ {
				msg.AddProgress()
				for k := range fem.params.Params {
					if fem.params.Params[k].Type == params.VolumeLoad {
						x := fem.mesh.FeCenter(j)
						if len(fem.params.Params[k].Predicate) > 0 {
							ok, err = fem.params.Params[k].GetPredicate(x, &fem.params.Variables)
							if err != nil {
								errChan <- err
								return
							}
							if !ok {
								continue
							}
						}
						value, err = fem.params.Params[k].GetValue(x, &fem.params.Variables)
						volume := fem.mesh.FeVolume(j)
						if err != nil {
							errChan <- err
							return
						}
						for l := 0; l < fem.mesh.FeSize(); l++ {
							data <- VectorData{index: fem.mesh.FE[j][l], direct: fem.params.Params[k].Direct, vector: [3]float64{volume * value * share[l], volume * value * share[l], volume * value * share[l]}}
						}
					}
				}
				atomic.AddInt32(&counter, 1)
				if int(counter) == fem.mesh.NumFE() {
					close(data)
				}
			}
		}()
	}
	if err := <-errChan; err != nil {
		return err
	}
	<-done
	return nil
}

func (fem *StaticFEM) addSurfaceLoad() error {
	if !fem.params.FindParameter(params.SurfaceLoad) && !fem.params.FindParameter(params.PressureLoad) {
		return nil
	}
	var share []float64
	switch fem.mesh.FeType {
	case mesh.Fe1d2:
		share = []float64{1.0}
	case mesh.Fe2d3:
		fallthrough
	case mesh.Fe2d4:
		share = []float64{0.5, 0.5}
	case mesh.Fe3d4:
		fallthrough
	case mesh.Fe3d3s:
		share = []float64{0.333333333333, 0.333333333333, 0.333333333333}
	case mesh.Fe3d8:
		fallthrough
	case mesh.Fe3d4s:
		share = []float64{0.25, 0.25, 0.25, 0.25}
	}
	var counter int32
	done := make(chan struct{})
	data := make(chan VectorData, fem.params.NumThread)
	errChan := make(chan error, fem.params.NumThread)
	step := fem.mesh.NumBE() / fem.params.NumThread
	go fem.addData(data, done, fem.solver.AddVector)
	msg := progress.NewProgress("Calculation of pressure/surface loads", 0, fem.mesh.NumBE(), 10)
	for i := 0; i < fem.params.NumThread; i++ {
		begin := i * step
		end := (i + 1) * step
		if i == fem.params.NumThread-1 {
			end = fem.mesh.NumBE()
		}
		go func() {
			var err error
			var ok bool
			var value float64
			defer func() {
				errChan <- err
			}()
			for j := begin; j < end; j++ {
				msg.AddProgress()
				for k := range fem.params.Params {
					if fem.params.Params[k].Type == params.SurfaceLoad || fem.params.Params[k].Type == params.PressureLoad {
						x := fem.mesh.BeCoord(j)
						if len(fem.params.Params[k].Predicate) > 0 {
							isValidPredicate := true
							for l := 0; l < fem.mesh.BeSize(); l++ {
								ok, err = fem.params.Params[k].GetPredicate(x.RowView(l).(*mat.VecDense), &fem.params.Variables)
								if err != nil {
									errChan <- err
									return
								}
								if !ok {
									isValidPredicate = false
									break
								}
							}
							if !isValidPredicate {
								continue
							}
						}
						value, err = fem.params.Params[k].GetValue(x.RowView(0).(*mat.VecDense), &fem.params.Variables)
						volume := fem.mesh.BeVolume(j)
						if err != nil {
							errChan <- err
							return
						}
						normal := [3]float64{1.0, 1.0, 1.0}
						if fem.params.Params[k].Type == params.PressureLoad {
							normal = fem.mesh.BeNormal(j)
						}
						for l := 0; l < fem.mesh.BeSize(); l++ {
							data <- VectorData{index: fem.mesh.BE[j][l], direct: fem.params.Params[k].Direct, vector: [3]float64{normal[0] * volume * value * share[l], normal[1] * volume * value * share[l], normal[2] * volume * value * share[l]}}
						}
					}
				}
				atomic.AddInt32(&counter, 1)
				if int(counter) == fem.mesh.NumBE() {
					close(data)
				}
			}
		}()
	}
	if err := <-errChan; err != nil {
		return err
	}
	<-done
	return nil
}

func (fem *StaticFEM) createFE(index int) (fe.FiniteElement, error) {
	cx := fem.mesh.FeCenter(index)
	x := fem.mesh.FeCoord(index)
	feParams := fe.FiniteElementParameters{}

	youngModulus, err := fem.params.GetParamValue(cx, params.YoungModulus)
	if err != nil {
		return nil, err
	}
	feParams.YoungModulus = youngModulus
	poissonRatio, err := fem.params.GetParamValue(cx, params.PoissonRatio)
	if err != nil {
		return nil, err
	}
	feParams.PoissonRation = poissonRatio

	if !fem.mesh.Is3D() {
		thickness, err := fem.params.GetParamValue(cx, params.Thickness)
		if err != nil {
			return nil, err
		}
		feParams.Thickness = thickness
	}

	switch fem.mesh.FeType {
	case mesh.Fe1d2:
		shape, err := fe.NewShape1d2(x)
		if err != nil {
			return nil, err
		}
		return fe.NewFE1D(shape, feParams), err
	case mesh.Fe2d3:
		shape, err := fe.NewShape2d3(x)
		if err != nil {
			return nil, err
		}
		return fe.NewFE2D(shape, feParams), err
	case mesh.Fe2d4:
		shape, err := fe.NewShape2d4(x)
		if err != nil {
			return nil, err
		}
		return fe.NewFE2D(shape, feParams), err
	case mesh.Fe3d4:
		shape, err := fe.NewShape3d4(x)
		if err != nil {
			return nil, err
		}
		return fe.NewFE3D(shape, feParams), err
	case mesh.Fe3d8:
		shape, err := fe.NewShape3d8(x)
		if err != nil {
			return nil, err
		}
		return fe.NewFE3D(shape, feParams), err
	case mesh.Fe3d3s:
		transformMatrix := util.TransformMatrix(x)
		shape, err := fe.NewShape2d3(util.Transpose(util.Mul(transformMatrix, x.T())))
		if err != nil {
			return nil, err
		}
		return fe.NewFE3DS(shape, transformMatrix, feParams), err
	case mesh.Fe3d4s:
		transformMatrix := util.TransformMatrix(x)
		shape, err := fe.NewShape2d4(util.Transpose(util.Mul(transformMatrix, x.T())))
		if err != nil {
			return nil, err
		}
		return fe.NewFE3DS(shape, transformMatrix, feParams), err
	}
	return nil, fmt.Errorf("bad finite element type")
}

func (fem *StaticFEM) numResult() int {
	var res int
	switch fem.mesh.FeType {
	case mesh.Fe1d2:
		res = 3 // U, Exx, Sxx
	case mesh.Fe2d3:
		fallthrough
	case mesh.Fe2d4:
		res = 8 // U, V, Exx, Eyy, Exy, Sxx, Syy, Sxy
	case mesh.Fe3d4:
		fallthrough
	case mesh.Fe3d8:
		res = 15 // U, V, W, Exx, Eyy, Ezz, Exy, Exz, Eyz, Sxx, Syy, Szz, Sxy, Sxz, Syz
	case mesh.Fe3d3s:
		fallthrough
	case mesh.Fe3d4s:
		res = 18 // U, V, W, Tx, Ty, Tz, Exx, Eyy, Ezz, Exy, Exz, Eyz, Sxx, Syy, Szz, Sxy, Sxz, Syz
	}
	return res
}

func (fem *StaticFEM) ResultNames() *[]string {
	var res []string
	switch fem.mesh.FeType {
	case mesh.Fe1d2:
		res = []string{"U", "Exx", "Sxx"}
	case mesh.Fe2d3:
		fallthrough
	case mesh.Fe2d4:
		res = []string{"U", "V", "Exx", "Eyy", "Exy", "Sxx", "Syy", "Sxy"}
	case mesh.Fe3d4:
		fallthrough
	case mesh.Fe3d8:
		res = []string{"U", "V", "W", "Exx", "Eyy", "Ezz", "Exy", "Exz", "Eyz", "Sxx", "Syy", "Szz", "Sxy", "Sxz", "Syz"}
	case mesh.Fe3d3s:
		fallthrough
	case mesh.Fe3d4s:
		res = []string{"U", "V", "W", "Tx", "Ty", "Tz", "Exx", "Eyy", "Ezz", "Exy", "Exz", "Eyz", "Sxx", "Syy", "Szz", "Sxy", "Sxz", "Syz"}
	}
	return &res
}

func (fem *StaticFEM) calcResult(u *mat.VecDense) error {
	var err error
	//var mt sync.Mutex
	fem.res = mat.NewDense(fem.numResult(), fem.mesh.NumVertex(), nil)
	counter := make([]int32, fem.mesh.NumVertex())
	data := make(chan MatrixData, fem.params.NumThread)
	errChan := make(chan error, fem.params.NumThread)
	done := make(chan struct{})
	// Copy the calculation save
	for i := 0; i < fem.mesh.NumVertex(); i++ {
		for j := 0; j < fem.mesh.Freedom(); j++ {
			fem.res.Set(j, i, u.AtVec(i*fem.mesh.Freedom()+j))
		}
	}
	// Calculation of standard save for all FE
	go func() {
		var local MatrixData
		msg := progress.NewProgress("Calculation of standard FE save", 0, fem.mesh.NumFE(), 10)
		for k := 0; k < fem.mesh.NumFE(); k++ {
			msg.AddProgress()
			local = <-data
			for i := 0; i < fem.numResult()-fem.mesh.Freedom(); i++ {
				for j := 0; j < fem.mesh.FeSize(); j++ {
					//mt.Lock()
					fem.res.Set(i+fem.mesh.Freedom(), fem.mesh.FE[local.index][j], fem.res.At(i+fem.mesh.Freedom(), fem.mesh.FE[local.index][j])+local.matrix.At(i, j))
					//mt.Unlock()
					if i == 0 {
						atomic.AddInt32(&counter[fem.mesh.FE[local.index][j]], 1)
					}
				}
			}
		}
		done <- struct{}{}
	}()

	step := fem.mesh.NumFE() / fem.params.NumThread
	for n := 0; n < fem.params.NumThread; n++ {
		begin := n * step
		end := (n + 1) * step
		if n == fem.params.NumThread-1 {
			end = fem.mesh.NumFE()
		}
		go func() {
			var elm fe.FiniteElement
			defer func() {
				errChan <- err
			}()
			for i := begin; i < end; i++ {
				if elm, err = fem.createFE(i); err != nil {
					errChan <- err
					return
				}
				// Form the displacement vector for the current FE
				feU := mat.NewVecDense(fem.mesh.FeSize()*fem.mesh.Freedom(), nil)
				for j := 0; j < fem.mesh.FeSize(); j++ {
					for k := 0; k < fem.mesh.Freedom(); k++ {
						feU.SetVec(j*fem.mesh.Freedom()+k, u.AtVec(fem.mesh.Freedom()*fem.mesh.FE[i][j]+k))
					}
				}
				feRes := elm.Calculate(feU)
				data <- MatrixData{index: i, matrix: feRes}
			}
		}()
	}
	if err = <-errChan; err != nil {
		return err
	}
	<-done
	// Average save
	for i := fem.mesh.Freedom(); i < fem.numResult(); i++ {
		for j := 0; j < fem.mesh.NumVertex(); j++ {
			fem.res.Set(i, j, fem.res.At(i, j)/float64(counter[j]))
			if math.Abs(fem.res.At(i, j)) < fem.params.Eps {
				fem.res.Set(i, j, 0)
			}
		}
	}
	return nil
}

//func (fem *StaticFEM) calcResult(u *mat.VecDense) error {
//	var err error
//	var wg sync.WaitGroup
//	var mt sync.Mutex
//	errChan := make(chan error, fem.params.NumThread)
//	fem.res = mat.NewDense(fem.numResult(), fem.mesh.NumVertex(), nil)
//	counter := make([]int32, fem.mesh.NumVertex())
//	// Copy the calculation save
//	for i := 0; i < fem.mesh.NumVertex(); i++ {
//		for j := 0; j < fem.mesh.Freedom(); j++ {
//			fem.res.Set(j, i, u.AtVec(i*fem.mesh.Freedom()+j))
//		}
//	}
//	// Calculation of standard save for all FE
//	msg := progress.NewProgress("Calculation of standard FE save", 0, fem.mesh.NumFE(), 10)
//	step := fem.mesh.NumFE() / fem.params.NumThread
//	wg.Add(fem.params.NumThread)
//	for l := 0; l < fem.params.NumThread; l++ {
//		begin := l * step
//		end := (l + 1) * step
//		if l == fem.params.NumThread-1 {
//			end = fem.mesh.NumFE()
//		}
//		go func() {
//			var elm fe.FiniteElement
//			defer func() {
//				wg.Done()
//				errChan <- err
//			}()
//			for i := begin; i < end; i++ {
//				msg.AddProgress()
//				if elm, err = fem.createFE(i); err != nil {
//					errChan <- err
//					return
//				}
//				// Create the displacement vector for the current FE
//				feU := mat.NewVecDense(fem.mesh.FeSize()*fem.mesh.Freedom(), nil)
//				for j := 0; j < fem.mesh.FeSize(); j++ {
//					for k := 0; k < fem.mesh.Freedom(); k++ {
//						feU.SetVec(j*fem.mesh.Freedom()+k, u.AtVec(fem.mesh.Freedom()*fem.mesh.FE[i][j]+k))
//					}
//				}
//				feRes := elm.Calculate(feU)
//				for m := 0; m < fem.numResult()-fem.mesh.Freedom(); m++ {
//					for j := 0; j < fem.mesh.FeSize(); j++ {
//						mt.Lock()
//						fem.res.Set(m+fem.mesh.Freedom(), fem.mesh.FE[i][j], fem.res.At(m+fem.mesh.Freedom(), fem.mesh.FE[i][j])+feRes.At(m, j))
//						mt.Unlock()
//						if m == 0 {
//							atomic.AddInt32(&counter[fem.mesh.FE[i][j]], 1)
//						}
//					}
//				}
//			}
//		}()
//	}
//	if err = <-errChan; err != nil {
//		return err
//	}
//	wg.Wait()
//	// Average save
//	for i := fem.mesh.Freedom(); i < fem.numResult(); i++ {
//		for j := 0; j < fem.mesh.NumVertex(); j++ {
//			fem.res.Set(i, j, fem.res.At(i, j)/float64(counter[j]))
//			if math.Abs(fem.res.At(i, j)) < fem.params.Eps {
//				fem.res.Set(i, j, 0)
//			}
//		}
//	}
//	//msg.StopProgress()
//	return nil
//}

func (fem *StaticFEM) SaveResult(name string) error {
	file, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("error creating result file")
	}
	defer func() {
		err = file.Close()
	}()
	// Signature
	if _, err = fmt.Fprintf(file, "FEM Solver Results File\n"); err != nil {
		return err
	}
	// Mesh
	if _, err = fmt.Fprintf(file, "Mesh\n"); err != nil {
		return err
	}
	if _, err = fmt.Fprintf(file, "%s\n", fem.mesh.FeName()); err != nil {
		return err
	}
	if _, err = fmt.Fprintf(file, "%d\n", fem.mesh.NumVertex()); err != nil {
		return err
	}
	for i := range fem.mesh.X {
		for j := range fem.mesh.X[i] {
			if _, err = fmt.Fprintf(file, "%f ", fem.mesh.X[i][j]); err != nil {
				return err
			}
		}
		if _, err = fmt.Fprintf(file, "\n"); err != nil {
			return err
		}
	}
	if _, err = fmt.Fprintf(file, "%d\n", fem.mesh.NumFE()); err != nil {
		return err
	}
	for i := range fem.mesh.FE {
		for j := range fem.mesh.FE[i] {
			if _, err = fmt.Fprintf(file, "%d ", fem.mesh.FE[i][j]); err != nil {
				return err
			}
		}
		if _, err = fmt.Fprintf(file, "\n"); err != nil {
			return err
		}
	}
	if _, err = fmt.Fprintf(file, "%d\n", fem.mesh.NumBE()); err != nil {
		return err
	}
	for i := range fem.mesh.BE {
		for j := range fem.mesh.BE[i] {
			if _, err = fmt.Fprintf(file, "%d ", fem.mesh.BE[i][j]); err != nil {
				return err
			}
		}
		if _, err = fmt.Fprintf(file, "\n"); err != nil {
			return err
		}
	}
	// Results
	if _, err = fmt.Fprintf(file, "Results\n"); err != nil {
		return err
	}
	now := time.Now()
	if _, err = fmt.Fprintf(file, "%02d.%02d.%4d - %02d:%02d:%02d\n", now.Day(), now.Month(), now.Year(), now.Hour(), now.Minute(), now.Second()); err != nil {
		return err
	}
	if _, err = fmt.Fprintf(file, "%d\n", fem.numResult()); err != nil {
		return err
	}
	rows, cols := fem.res.Dims()
	for i := 0; i < rows; i++ {
		if _, err = fmt.Fprintf(file, "%s\n", (*fem.ResultNames())[i]); err != nil {
			return err
		}
		if _, err = fmt.Fprintf(file, "0\n%d\n", cols); err != nil {
			return err
		}
		for j := 0; j < cols; j++ {
			if _, err = fmt.Fprintf(file, "%0.8e\n", fem.res.At(i, j)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (fem *StaticFEM) GetMesh() *mesh.Mesh {
	return &fem.mesh
}
