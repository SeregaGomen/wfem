package web

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"wfem/cmd/fem/params"
)

type condition struct {
	Value, Predicate, raw string
	Direction             int
}

type result struct {
	Name     string
	Min, Max float64
}

type report struct {
	DateTime, FeName, Mesh                                                                          string
	NumFE, NumVertex                                                                                int
	Variables                                                                                       map[string]float64
	YoungModulus, PoissonRatio, VolumeLoad, SurfaceLoad, PointLoad, PressureLoad, BoundaryCondition []condition
	Res                                                                                             []result
}

type problemInfo struct {
	Mesh                                                                                                                  []string
	YoungModulus, PoissonRatio, Thickness, VolumeLoad, SurfaceLoad, PointLoad, PressureLoad, BoundaryCondition, Variables string
	Threads                                                                                                               int
	Eps                                                                                                                   float64
}

var tmpl *template.Template

// var problem problemInfo
var rep report

func init() {
	tmpl = template.New("allTemplates")
	tmpl.Funcs(template.FuncMap{
		"isX": func(dir int) string {
			if dir&params.X == params.X {
				return "+"
			}
			return ""
		},
		"isY": func(dir int) string {
			if dir&params.Y == params.Y {
				return "+"
			}
			return ""
		},
		"isZ": func(dir int) string {
			if dir&params.Z == params.Z {
				return "+"
			}
			return ""
		}})
	if _, err := tmpl.ParseGlob("ui/html/*.*"); err != nil {
		log.Fatal("500 Internal Server Error", err)
	}
	//problem = problemInfo{Eps: 1.0e-10, Threads: runtime.NumCPU()}
}

func newProblemPageHandler(writer http.ResponseWriter, request *http.Request) {
	var err error
	if request.URL.Path != "/problem/" {
		http.NotFound(writer, request)
		return
	}
	if err = request.ParseForm(); err != nil {
		log.Fatal("500 Internal Server Error: ", err)
	}
	problem := problemInfo{Eps: 1.0e-10, Threads: runtime.NumCPU()}

	problemName := request.FormValue("problem")
	if len(problemName) > 0 {
		if err = loadJson("save/"+problemName, &problem); err != nil {
			log.Fatal("500 Internal Server Error: ", err)
		}
	} else if problem.Mesh, err = scanDir("downloads"); err != nil {
		log.Fatal("500 Internal Server Error: ", err)
	}
	if err = tmpl.ExecuteTemplate(writer, "problem.html", &problem); err != nil {
		log.Fatal("500 Internal Server Error: ", err)
	}
}

func scanDir(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, file := range files {
		results = append(results, file.Name())
	}
	return results, nil
}

func strCondition(cond *[]condition) string {
	var str string
	for i := range *cond {
		str += (*cond)[i].raw + "\n"
	}
	return str
}

func strVariables(variables *map[string]float64) string {
	var str string
	for name, value := range *variables {
		str += name + "=" + strconv.FormatFloat(value, 'f', -1, 64) + "\n"
	}
	return str
}

func saveJson(data *problemInfo) error {
	fileName := "save/" + (*data).Mesh[0][0:len((*data).Mesh[0])-len(filepath.Ext((*data).Mesh[0]))] + ".json"
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(fileName, file, 0644)
}

func loadJson(fileName string, data *problemInfo) error {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, &data)
}

func loadProblemPageHandler(writer http.ResponseWriter, request *http.Request) {
	var err error
	var problems []string
	if request.URL.Path != "/load/" {
		http.NotFound(writer, request)
		return
	}
	if err = request.ParseForm(); err != nil {
		log.Fatal("500 Internal Server Error: ", err)
	}
	if problems, err = scanDir("save"); err != nil {
		log.Fatal("500 Internal Server Error: ", err)
	}
	if err = tmpl.ExecuteTemplate(writer, "load.html", &problems); err != nil {
		log.Fatal("500 Internal Server Error: ", err)
	}
	//if err = loadProblemProcessRequest(writer, request); err != nil {
	//	alert(writer, err)
	//}
}
