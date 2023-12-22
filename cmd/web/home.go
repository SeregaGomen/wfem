package web

import (
	"log"
	"net/http"
)

func homePageHandler(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		http.NotFound(writer, request)
		return
	}

	if err := request.ParseForm(); err != nil {
		log.Fatal("500 Internal Server Error: ", err)
	}
	if err := tmpl.ExecuteTemplate(writer, "home.html", nil); err != nil {
		log.Fatal("500 Internal Server Error: ", err)
	}
}
