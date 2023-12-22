package web

import (
	"log"
	"net/http"
	"strconv"
)

func StartServer(port int) {
	http.HandleFunc("/", homePageHandler)
	http.HandleFunc("/mesh/", meshPageHandler)
	http.HandleFunc("/info/", meshInfoPageHandler)
	http.HandleFunc("/problem/", newProblemPageHandler)
	http.HandleFunc("/load/", loadProblemPageHandler)
	http.HandleFunc("/results/", resultsPageHandler)

	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./ui/static"))))
	http.Handle("/downloads/", http.StripPrefix("/downloads", http.FileServer(http.Dir("./downloads"))))
	http.Handle("/save/", http.StripPrefix("/save", http.FileServer(http.Dir("./save"))))

	log.Println("Running a WebFEM server on localhost:" + strconv.Itoa(port))
	log.Fatal("failed to start server", http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

// https://webfem.sourceforge.net/
// https://golangify.com/
