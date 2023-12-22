package web

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wfem/cmd/fem/fem"
	"wfem/cmd/fem/params"
)

func resultsPageHandler(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/results/" {
		http.NotFound(writer, request)
		return
	}
	if err := request.ParseForm(); err != nil {
		log.Fatal("500 Internal Server Error: ", err)
	} else if err = resultProcessRequest(writer, request); err != nil {
		alert(writer, err)
	} else if err = tmpl.ExecuteTemplate(writer, "results.html", &rep); err != nil {
		log.Fatal("500 Internal Server Error: ", err)
	}
}

func getParam(request *http.Request, paramName string, cond *[]condition, isDirect bool) (bool, error) {
	var (
		ok        bool
		err       error
		value     string
		predicate string
		raw       string
		direct    int
	)
	if fields, found := request.Form[paramName]; found && len(fields) > 0 {
		*cond = make([]condition, 0, 10)
		field := strings.Split(fields[0], "\n")
		for i := range field {
			if len(strings.Trim(field[i], " ")) == 0 {
				continue
			}
			data := strings.Trim(field[i], "\n\r\t")
			raw = data
			val := strings.Split(data, ";")
			value = val[0]
			if len(val) > 1 {
				predicate = val[1]
			} else {
				predicate = ""
			}
			if len(val) > 2 && isDirect {
				if direct, err = func(dirStr string) (int, error) {
					dir := 0
					val := strings.Split(dirStr, "|")
					for i := range val {
						switch val[i] {
						case "X":
							dir |= params.X
						case "Y":
							dir |= params.Y
						case "Z":
							dir |= params.Z
						default:
							return dir, fmt.Errorf("invalid direction")
						}
					}
					return dir, nil
				}(val[2]); err != nil {
					return false, err
				}
			} else {
				direct = 0
			}
			*cond = append(*cond, condition{Value: value, Predicate: predicate, raw: raw, Direction: direct})
			ok = true
		}
		return ok, nil
	}
	return ok, nil
}

func resultProcessRequest(_ http.ResponseWriter, request *http.Request) error {
	var (
		numThreads                                                                                                 int
		eps                                                                                                        float64
		meshName                                                                                                   string
		thickness, youngModulus, poissonRatio, volumeLoad, pointLoad, surfaceLoad, pressureLoad, boundaryCondition []condition
		variables                                                                                                  map[string]float64
	)

	// Mesh
	if meshName = request.FormValue("mesh"); len(meshName) == 0 {
		return fmt.Errorf("wrong mesh file name")
	}
	// Threads
	field := request.FormValue("threads")
	if x, err := strconv.Atoi(field); err != nil {
		return fmt.Errorf("parameter 'Number of thread' is invalid")
	} else {
		numThreads = x
	}
	// Tolerance
	field = request.FormValue("eps")
	if x, err := strconv.ParseFloat(field, 64); err != nil {
		return fmt.Errorf("parameter 'Tolerance' is invalid")
	} else {
		eps = x
	}
	// Variables
	fields := strings.Split(request.FormValue("variables"), "\n")
	variables = map[string]float64{}
	for i := range fields {
		if len(strings.Trim(fields[i], " ")) == 0 {
			continue
		}
		val := strings.Split(strings.Trim(fields[i], " \n\r\t"), "=")
		name := strings.Trim(val[0], " \n\r\t")
		if len(val) == 2 {
			value, err := strconv.ParseFloat(strings.Trim(val[1], " \n\r\t"), 64)
			if err != nil {
				return err
			}
			variables[name] = value
		} else {
			return fmt.Errorf("invalid variable")
		}
	}

	// Young modulus
	if ok, err := getParam(request, "young_modulus", &youngModulus, false); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("parameter 'Young modulus' is invalid")
	}
	// Poisson's ratio
	if ok, err := getParam(request, "poisson_ratio", &poissonRatio, false); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("parameter 'Poisson's ratio' is invalid")
	}
	// Thickness
	if ok, err := getParam(request, "thickness", &thickness, false); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("parameter 'Thickness' is invalid")
	}
	// Volume load
	if _, err := getParam(request, "volume_load", &volumeLoad, true); err != nil {
		return err
	}
	// Point load
	if _, err := getParam(request, "point_load", &pointLoad, true); err != nil {
		return err
	}
	// Surface load
	if _, err := getParam(request, "surface_load", &surfaceLoad, true); err != nil {
		return err
	}
	// Pressure load
	if _, err := getParam(request, "pressure_load", &pressureLoad, false); err != nil {
		return err
	}
	// Boundary conditions
	if ok, err := getParam(request, "boundary_condition", &boundaryCondition, true); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("parameter 'Boundary condition' is invalid")
	}

	f := fem.NewStaticFEM()
	err := f.SetMesh("downloads/" + meshName)
	if err != nil {
		return err
	}
	f.SetNumThread(numThreads)
	f.SetEps(eps)
	for i := range youngModulus {
		f.AddYoungModulus(youngModulus[i].Value, youngModulus[i].Predicate)
	}
	for i := range poissonRatio {
		f.AddPoissonRatio(poissonRatio[i].Value, poissonRatio[i].Predicate)
	}
	for i := range thickness {
		f.AddThickness(thickness[i].Value, thickness[i].Predicate)
	}
	for i := range volumeLoad {
		f.AddVolumeLoad(volumeLoad[i].Value, volumeLoad[i].Predicate, volumeLoad[i].Direction)
	}
	for i := range surfaceLoad {
		f.AddSurfaceLoad(surfaceLoad[i].Value, surfaceLoad[i].Predicate, surfaceLoad[i].Direction)
	}
	for i := range pointLoad {
		f.AddPointLoad(pointLoad[i].Value, pointLoad[i].Predicate, pointLoad[i].Direction)
	}
	for i := range pressureLoad {
		f.AddPressureLoad(pressureLoad[i].Value, pressureLoad[i].Predicate)
	}
	for i := range boundaryCondition {
		f.AddBoundaryCondition(boundaryCondition[i].Value, boundaryCondition[i].Predicate, boundaryCondition[i].Direction)
	}
	for name, value := range variables {
		f.AddVariable(name, value)
	}

	problem := problemInfo{Mesh: []string{meshName}, Threads: numThreads, Eps: eps, YoungModulus: strCondition(&youngModulus),
		PoissonRatio: strCondition(&poissonRatio), Thickness: strCondition(&thickness),
		VolumeLoad: strCondition(&volumeLoad), SurfaceLoad: strCondition(&surfaceLoad),
		PointLoad: strCondition(&pointLoad), PressureLoad: strCondition(&pressureLoad),
		BoundaryCondition: strCondition(&boundaryCondition), Variables: strVariables(&variables),
	}
	if err = saveJson(&problem); err != nil {
		return err
	}
	if err = f.Calculate(); err != nil {
		return err
	}
	if res := func(res *mat.Dense, names *[]string) []result {
		r := make([]result, 0, len(*names))
		for i, name := range *names {
			r = append(r, result{Name: name, Min: mat.Min(res.RowView(i)), Max: mat.Max(res.RowView(i))})
		}
		return r
	}(f.GetResult(), f.ResultNames()); res != nil {
		rep = report{DateTime: time.Now().Format("01-02-2006 15:04:05"), FeName: f.GetMesh().FeName(),
			NumFE: f.GetMesh().NumFE(), NumVertex: f.GetMesh().NumVertex(), YoungModulus: youngModulus,
			PoissonRatio: poissonRatio, VolumeLoad: volumeLoad, SurfaceLoad: surfaceLoad, PointLoad: pointLoad,
			PressureLoad: pressureLoad, BoundaryCondition: boundaryCondition, Variables: variables, Mesh: problem.Mesh[0],
			Res: res}
	}
	return nil
}

func alert(writer http.ResponseWriter, err error) {
	//_, _ = fmt.Fprintf(writer, `<p class="error">Error: %s</p>`, err)
	//_, _ = fmt.Fprintf(writer, `<script>alert("Error: %s")</script>`, err.Error())
	_, _ = fmt.Fprintf(writer, `<script>alert("Error: %s");window.history.back();</script>`, err.Error())
}
