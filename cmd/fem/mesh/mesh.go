package mesh

import (
	"bufio"
	"fmt"
	"golang.org/x/exp/slices"
	"gonum.org/v1/gonum/mat"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"wfem/cmd/fem/progress"
	"wfem/cmd/fem/util"
)

const (
	Fe1d2 int = iota
	Fe2d3
	Fe2d4
	Fe3d4
	Fe3d8
	Fe3d3s
	Fe3d4s
	// Is1D2(), ...
)

type Mesh struct {
	FeType  int
	X       [][]float64
	FE      [][]int
	BE      [][]int
	MeshMap [][]int
}

func (m *Mesh) NumVertex() int {
	return len(m.X)
}

func (m *Mesh) NumFE() int {
	return len(m.FE)
}

func (m *Mesh) NumBE() int {
	return len(m.BE)
}

func (m *Mesh) Freedom() int {
	_, _, _, freedom, _ := feParam(m.FeType)
	return freedom
}

func (m *Mesh) Nnz() int {
	nnz := 0
	for i := range m.MeshMap {
		nnz += len(m.MeshMap[i])
	}
	return nnz
}

func (m *Mesh) Load(name string) error {
	var err error
	switch strings.ToUpper(filepath.Ext(name)) {
	case ".VOL": // Netgen
		err = m.loadVol(name)
	case ".MSH": // Gmsh
		err = m.loadMsh(name)
	case ".MESH":
		err = m.loadMesh(name)
	default:
		return fmt.Errorf("wrong mesh-file format")
	}
	if err == nil {
		fmt.Println("Mesh file:", name)
		fmt.Println("Finite element type:", m.FeName())
		fmt.Println("Number of nodes:", m.NumVertex())
		fmt.Println("Number of finite element:", len(m.FE))
		m.CreateMeshMap()
	}
	return err
}

func (m *Mesh) loadVol(name string) error {
	var num int
	var data []string
	file, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("error opening file")
	}
	defer func() {
		err = file.Close()
	}()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if scanner.Text() == "surfaceelements" {
			break
		}
	}
	if scanner.Text() == "" {
		return fmt.Errorf("wrong VOL-file format")
	}

	m.FeType = Fe3d4
	// Num of boundary elements
	scanner.Scan()
	if num, err = strconv.Atoi(scanner.Text()); err != nil {
		return err
	}

	m.BE = make([][]int, num)
	for i := range m.BE {
		m.BE[i] = make([]int, 3)
		scanner.Scan()
		data = strings.Fields(scanner.Text())
		// data = strings.Split(strings.Trim(scanner.Text(), " "), " ")
		for j := 0; j < 3; j++ {
			m.BE[i][j], err = strconv.Atoi(data[j+5])
			if err != nil {
				return err
			}
			m.BE[i][j] -= 1
		}
	}

	for scanner.Scan() {
		if scanner.Text() == "volumeelements" {
			break
		}
	}
	if scanner.Text() == "" {
		return fmt.Errorf("wrong VOL-file format")
	}

	// Num of finite elements
	scanner.Scan()
	if num, err = strconv.Atoi(scanner.Text()); err != nil {
		return err
	}

	m.FE = make([][]int, num)
	for i := range m.FE {
		m.FE[i] = make([]int, 4)
		scanner.Scan()
		data = strings.Fields(scanner.Text())
		// data = strings.Split(strings.Trim(scanner.Text(), " "), " ")
		for j := 0; j < 4; j++ {
			m.FE[i][j], err = strconv.Atoi(data[j+2])
			if err != nil {
				return err
			}
			m.FE[i][j] -= 1
		}
	}

	for scanner.Scan() {
		if scanner.Text() == "points" {
			break
		}
	}
	if scanner.Text() == "" {
		return fmt.Errorf("wrong VOL-file format")
	}

	// Coordinates
	scanner.Scan()
	if num, err = strconv.Atoi(scanner.Text()); err != nil {
		return err
	}

	m.X = make([][]float64, num)
	for i := range m.X {
		m.X[i] = make([]float64, 3)
		scanner.Scan()
		data = strings.Fields(scanner.Text())
		for j := 0; j < 3; j++ {
			m.X[i][j], err = strconv.ParseFloat(data[j], 16)
			if err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

func feParam(feType int) (beSize, feSize, feDim, freedom int, err error) {
	err = nil
	switch feType {
	case Fe1d2:
		beSize = 1
		feSize = 2
		feDim = 1
		freedom = 1
	case Fe2d3:
		beSize = 2
		feSize = 3
		feDim = 2
		freedom = 2
	case Fe2d4:
		beSize = 2
		feSize = 4
		feDim = 2
		freedom = 2
	case Fe3d4:
		beSize = 3
		feSize = 4
		feDim = 3
		freedom = 3
	case Fe3d8:
		beSize = 4
		feSize = 8
		feDim = 3
		freedom = 3
	case Fe3d3s:
		beSize = 3
		feSize = 3
		feDim = 3
		freedom = 6
	case Fe3d4s:
		beSize = 4
		feSize = 4
		feDim = 3
		freedom = 6
	default:
		err = fmt.Errorf("unknown FE type")
	}
	return beSize, feSize, feDim, freedom, err
}

func (m *Mesh) loadMesh(name string) error {
	var num int
	var val string
	file, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("error opening file")
	}
	defer func() {
		err = file.Close()
	}()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	// Mesh type
	scanner.Scan()
	val = scanner.Text()
	switch val {
	case "fe1d2":
		m.FeType = Fe1d2
	case "fe2d3":
		m.FeType = Fe2d3
	case "fe2d4":
		m.FeType = Fe2d4
	case "fe3d4":
		m.FeType = Fe3d4
	case "fe3d8":
		m.FeType = Fe3d8
	case "fe3d3s":
		m.FeType = Fe3d3s
	case "fe3d4s":
		m.FeType = Fe3d4s
	default:
		return fmt.Errorf("unknown FE type")
	}
	beSize, feSize, feDim, _, err := feParam(m.FeType)
	if err != nil {
		return fmt.Errorf("wrong MESH-file format")
	}

	// Coordinates
	scanner.Scan()
	val = scanner.Text()
	if num, err = strconv.Atoi(val); err != nil {
		return err
	}

	m.X = make([][]float64, num)
	for i := range m.X {
		m.X[i] = make([]float64, feDim)
		for j := 0; j < feDim; j++ {
			scanner.Scan()
			val = scanner.Text()
			m.X[i][j], err = strconv.ParseFloat(val, 16)
			if err != nil {
				return err
			}
		}
	}
	// Finite elements
	scanner.Scan()
	val = scanner.Text()
	if num, err = strconv.Atoi(val); err != nil {
		return err
	}
	m.FE = make([][]int, num)
	for i := range m.FE {
		m.FE[i] = make([]int, feSize)
		for j := 0; j < feSize; j++ {
			scanner.Scan()
			val = scanner.Text()
			m.FE[i][j], err = strconv.Atoi(val)
			if err != nil {
				return err
			}
		}
	}
	// Boundary elements
	scanner.Scan()
	val = scanner.Text()
	if num, err = strconv.Atoi(val); err != nil {
		return err
	}
	m.BE = make([][]int, num)
	for i := range m.BE {
		m.BE[i] = make([]int, beSize)
		for j := 0; j < beSize; j++ {
			scanner.Scan()
			val = scanner.Text()
			m.BE[i][j], err = strconv.Atoi(val)
			if err != nil {
				return err
			}
		}
	}
	if m.FeType == Fe3d3s || m.FeType == Fe3d4s {
		m.BE = m.FE
	}
	return nil
}

func (m *Mesh) loadMsh(name string) error {
	var data []string
	var numEntities, num, dim, minTag, elmType int
	is2d := true
	eps := 1.0e-10

	file, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("error opening file")
	}
	defer func() {
		err = file.Close()
	}()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	scanner.Scan()
	if scanner.Text() != "$MeshFormat" {
		return fmt.Errorf("wrong MSH-file format")
	}

	for scanner.Scan() {
		if scanner.Text() == "$Nodes" {
			break
		}
	}
	if scanner.Text() == "" {
		return fmt.Errorf("wrong MSH-file format")
	}

	// Number of section
	scanner.Scan()

	data = strings.Split(scanner.Text(), " ")
	if numEntities, err = strconv.Atoi(data[0]); err != nil {
		return err
	}
	if num, err = strconv.Atoi(data[1]); err != nil {
		return err
	}
	m.X = make([][]float64, 0, num)
	for i := 0; i < numEntities; i++ {
		scanner.Scan()
		// data = strings.Split(scanner.Text(), " ")
		data = strings.Fields(scanner.Text())
		if num, err = strconv.Atoi(data[3]); err != nil {
			return err
		}
		// Ignoring tags
		for j := 0; j < num; j++ {
			scanner.Scan()
		}
		for j := 0; j < num; j++ {
			x := make([]float64, 3)
			scanner.Scan()
			// data = strings.Split(scanner.Text(), " ")
			data = strings.Fields(scanner.Text())
			for k := 0; k < 3; k++ {
				x[k], err = strconv.ParseFloat(data[k], 16)
				if err != nil {
					return err
				}
			}
			if math.Abs(x[2]) > eps {
				is2d = false
			}
			m.X = append(m.X, x)
		}
	}
	scanner.Scan()
	if scanner.Text() != "$EndNodes" {
		return fmt.Errorf("wrong MSH-file format")
	}

	scanner.Scan()
	if scanner.Text() != "$Elements" {
		return fmt.Errorf("wrong MSH-file format")
	}

	// Number of section
	scanner.Scan()
	// data = strings.Split(scanner.Text(), " ")
	data = strings.Fields(scanner.Text())
	if numEntities, err = strconv.Atoi(data[0]); err != nil {
		return err
	}
	if num, err = strconv.Atoi(data[1]); err != nil {
		return err
	}
	if minTag, err = strconv.Atoi(data[2]); err != nil {
		return err
	}

	m.BE = make([][]int, 0, num)
	m.FE = make([][]int, 0, num)

	for i := 0; i < numEntities; i++ {
		scanner.Scan()
		// data = strings.Split(scanner.Text(), " ")
		data = strings.Fields(scanner.Text())
		if dim, err = strconv.Atoi(data[0]); err != nil {
			return err
		}
		if elmType, err = strconv.Atoi(data[2]); err != nil {
			return err
		}
		if num, err = strconv.Atoi(data[3]); err != nil {
			return err
		}

		for j := 0; j < num; j++ {
			scanner.Scan()
			if dim == 0 || (dim == 1 && is2d == false) {
				continue
			}
			data = strings.Fields(scanner.Text())
			// data = strings.Split(strings.Trim(scanner.Text(), " "), " ")

			// Reading current element
			elm := make([]int, len(data)-1)
			for k := 1; k < len(data); k++ {
				elm[k-1], err = strconv.Atoi(data[k])
				if err != nil {
					return err
				}
				elm[k-1] -= minTag
			}

			switch elmType {
			case 1: // 2-node line
				if is2d == true {
					// Boundary element
					m.BE = append(m.BE, elm)
				}
			case 2: // 3-node triangle
				if is2d == true {
					// Finite element
					m.FE = append(m.FE, elm)
				} else {
					// Boundary element
					m.BE = append(m.BE, elm)
				}
			case 4: // 4-node tetrahedron
				if is2d == false {
					// Finite element
					m.FE = append(m.FE, elm)
				} else {
					return fmt.Errorf("this format of MSH-file is not supported")
				}
			default:
				return fmt.Errorf("this format of MSH-file is not supported")
			}
		}
	}
	scanner.Scan()
	if scanner.Text() != "$EndElements" {
		return fmt.Errorf("wrong MSH-file format")
	}

	if is2d == true {
		m.FeType = Fe2d3
	} else {
		m.FeType = Fe3d4
		if len(m.FE) == 0 {
			// Shell finite element
			m.FE = m.BE
			m.BE = m.BE[:0:0]
			m.FeType = Fe3d3s
		}
	}

	// Shrink to fit
	m.BE = m.BE[:len(m.BE):len(m.BE)]
	m.FE = m.FE[:len(m.FE):len(m.FE)]

	//m.Save("d:/cube.mesh")
	return nil
}

func (m *Mesh) FeName() string {
	ret := ""
	switch m.FeType {
	case Fe1d2:
		ret = "fe1d2"
	case Fe2d3:
		ret = "fe2d3"
	case Fe2d4:
		ret = "fe2d4"
	case Fe3d4:
		ret = "fe3d4"
	case Fe3d8:
		ret = "fe3d8"
	case Fe3d3s:
		ret = "fe3d3s"
	case Fe3d4s:
		ret = "fe3d4s"
	}
	return ret
}

func (m *Mesh) Save(name string) error {
	file, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("error opening file")
	}
	defer func() {
		err = file.Close()
	}()
	_, err = fmt.Fprintf(file, "%s\n", m.FeName())
	if err != nil {
		return fmt.Errorf("error writing MESH-file")
	}
	_, err = fmt.Fprintf(file, "%d\n", m.NumVertex())
	if err != nil {
		return fmt.Errorf("error writing MESH-file")
	}
	_, _, feDim, _, _ := feParam(m.FeType)
	for i := range m.X {
		for j := 0; j < feDim; j++ {
			_, err = fmt.Fprintf(file, "%f ", m.X[i][j])
			if err != nil {
				return fmt.Errorf("error writing MESH-file")
			}
		}
		_, err = fmt.Fprintln(file)
		if err != nil {
			return fmt.Errorf("error writing MESH-file")
		}
	}
	_, err = fmt.Fprintf(file, "%d\n", len(m.FE))
	if err != nil {
		return fmt.Errorf("error writing MESH-file")
	}
	for i := range m.FE {
		for j := 0; j < len(m.FE[i]); j++ {
			_, err = fmt.Fprintf(file, "%d ", m.FE[i][j])
			if err != nil {
				return fmt.Errorf("error writing MESH-file")
			}
		}
		_, err = fmt.Fprintln(file)
		if err != nil {
			return fmt.Errorf("error writing MESH-file")
		}
	}
	_, err = fmt.Fprintf(file, "%d\n", len(m.BE))
	if err != nil {
		return fmt.Errorf("error writing MESH-file")
	}
	for i := range m.BE {
		for j := 0; j < len(m.BE[i]); j++ {
			_, err = fmt.Fprintf(file, "%d ", m.BE[i][j])
			if err != nil {
				return fmt.Errorf("error writing MESH-file")
			}
		}
		_, err = fmt.Fprintln(file)
		if err != nil {
			return fmt.Errorf("error writing MESH-file")
		}
	}
	return nil
}

func (m *Mesh) BeNormal(index int) [3]float64 {
	var norm [3]float64
	if m.Is1D() {
		norm = [3]float64{1.0, 0.0, 0.0}
	} else if m.Is2D() {
		norm = [3]float64{
			m.X[m.BE[index][0]][1] - m.X[m.BE[index][1]][1],
			m.X[m.BE[index][1]][0] - m.X[m.BE[index][0]][0],
			0.0,
		}
	} else {
		norm = [3]float64{
			(m.X[m.BE[index][1]][1]-m.X[m.BE[index][0]][1])*(m.X[m.BE[index][2]][2]-m.X[m.BE[index][0]][2]) - (m.X[m.BE[index][2]][1]-m.X[m.BE[index][0]][1])*(m.X[m.BE[index][1]][2]-m.X[m.BE[index][0]][2]),
			(m.X[m.BE[index][2]][0]-m.X[m.BE[index][0]][0])*(m.X[m.BE[index][1]][2]-m.X[m.BE[index][0]][2]) - (m.X[m.BE[index][1]][0]-m.X[m.BE[index][0]][0])*(m.X[m.BE[index][2]][2]-m.X[m.BE[index][0]][2]),
			(m.X[m.BE[index][1]][0]-m.X[m.BE[index][0]][0])*(m.X[m.BE[index][2]][1]-m.X[m.BE[index][0]][1]) - (m.X[m.BE[index][2]][0]-m.X[m.BE[index][0]][0])*(m.X[m.BE[index][1]][1]-m.X[m.BE[index][0]][1]),
		}
	}
	l := math.Sqrt(norm[0]*norm[0] + norm[1]*norm[1] + norm[2]*norm[2])
	norm[0] /= l
	norm[1] /= l
	norm[2] /= l
	return norm
}

func (m *Mesh) BeCoord(index int) *mat.Dense {
	beSize, _, feDim, _, _ := feParam(m.FeType)
	res := make([]float64, beSize*feDim)
	for i := 0; i < beSize; i++ {
		for j := 0; j < feDim; j++ {
			res[i*feDim+j] = m.X[m.BE[index][i]][j]
		}
	}
	return mat.NewDense(beSize, feDim, res)
}

func (m *Mesh) FeCoord(index int) *mat.Dense {
	_, feSize, feDim, _, _ := feParam(m.FeType)
	res := make([]float64, feSize*feDim)
	for i := 0; i < feSize; i++ {
		for j := 0; j < feDim; j++ {
			res[i*feDim+j] = m.X[m.FE[index][i]][j]
		}
	}
	return mat.NewDense(feSize, feDim, res)
}

func (m *Mesh) FeCenter(index int) *mat.VecDense {
	_, feSize, feDim, _, _ := feParam(m.FeType)
	xc := mat.NewVecDense(feDim, nil)
	// Center of finite element
	for j := 0; j < feDim; j++ {
		c := 0.0
		for i := 0; i < feSize; i++ {
			c += m.X[m.FE[index][i]][j]
		}
		xc.SetVec(j, c/float64(feSize))
	}
	return xc
}

func (m *Mesh) BeCenter(index int) *mat.VecDense {
	beSize, _, feDim, _, _ := feParam(m.FeType)
	xc := mat.NewVecDense(feDim, nil)
	// Center of boundary element
	for j := 0; j < feDim; j++ {
		c := 0.0
		for i := 0; i < beSize; i++ {
			c += m.X[m.BE[index][i]][j]
		}
		xc.SetVec(j, c/float64(beSize))
	}
	return xc
}

func (m *Mesh) FeSize() int {
	_, feSize, _, _, _ := feParam(m.FeType)
	return feSize
}

func (m *Mesh) BeSize() int {
	beSize, _, _, _, _ := feParam(m.FeType)
	return beSize
}

func (m *Mesh) FeDim() int {
	_, _, feDim, _, _ := feParam(m.FeType)
	return feDim
}

func (m *Mesh) Is1D() bool {
	if m.FeType == Fe1d2 {
		return true
	}
	return false
}

func (m *Mesh) Is2D() bool {
	if m.FeType == Fe2d3 || m.FeType == Fe2d4 {
		return true
	}
	return false
}

func (m *Mesh) Is3D() bool {
	if m.FeType == Fe3d4 || m.FeType == Fe3d8 {
		return true
	}
	return false
}

func (m *Mesh) IsShell() bool {
	if m.FeType == Fe3d3s || m.FeType == Fe3d4s {
		return true
	}
	return false
}

func (m *Mesh) BeVolume(index int) float64 {
	var res float64
	x := m.BeCoord(index)
	switch m.FeType {
	case Fe1d2:
		res = 1.0
	case Fe2d3:
		fallthrough
	case Fe2d4:
		res = util.Volume1d2(x)
	case Fe3d4:
		fallthrough
	case Fe3d3s:
		res = util.Volume2d3(x)
	case Fe3d8:
		fallthrough
	case Fe3d4s:
		res = util.Volume2d4(x)
	}
	return res
}

func (m *Mesh) FeVolume(index int) float64 {
	var res float64
	x := m.FeCoord(index)
	switch m.FeType {
	case Fe1d2:
		res = util.Volume1d2(x)
	case Fe2d3:
		fallthrough
	case Fe3d3s:
		res = util.Volume2d3(x)
	case Fe2d4:
		fallthrough
	case Fe3d4s:
		res = util.Volume2d4(x)
	case Fe3d4:
		res = util.Volume3d4(x)
	case Fe3d8:
		res = util.Volume3d8(x)
	}
	return res
}

//func (m *Mesh) CreateMeshMap() {
//	m.MeshMap = make([][]int, m.NumVertex())
//	msg := progress.NewProgress("Creating mesh map", 0, m.NumFE(), 10)
//	for i := range m.FE {
//		msg.AddProgress()
//		for j := range m.FE[i] {
//			for k := range m.FE[i] {
//				if !slices.Contains(m.MeshMap[m.FE[i][j]], m.FE[i][k]) {
//					m.MeshMap[m.FE[i][j]] = append(m.MeshMap[m.FE[i][j]], m.FE[i][k])
//				}
//			}
//		}
//	}
//	for i := 0; i < m.NumVertex(); i++ {
//		sort.Ints(m.MeshMap[i])
//	}
//}

func (m *Mesh) CreateMeshMap() {
	m.MeshMap = make([][]int, m.NumVertex())
	msg := progress.NewProgress("Creating mesh map", 0, m.NumFE(), 10)
	for i := range m.FE {
		msg.AddProgress()
		for j := range m.FE[i] {
			for k := range m.FE[i] {
				if m.FE[i][k] < m.FE[i][j] {
					continue
				}
				if !slices.Contains(m.MeshMap[m.FE[i][j]], m.FE[i][k]) {
					m.MeshMap[m.FE[i][j]] = append(m.MeshMap[m.FE[i][j]], m.FE[i][k])
				}
			}
		}
	}
	for i := 0; i < m.NumVertex(); i++ {
		sort.Ints(m.MeshMap[i])
	}
}
