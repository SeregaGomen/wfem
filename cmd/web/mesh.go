package web

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

func meshPageHandler(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/mesh/" {
		http.NotFound(writer, request)
		return
	}
	if err := request.ParseForm(); err != nil {
		//if err := request.ParseMultipartForm(100000); err != nil {
		log.Fatal("500 Internal Server Error: ", err)
	}
	if err := tmpl.ExecuteTemplate(writer, "mesh.html", nil); err != nil {
		log.Fatal("500 Internal Server Error: ", err)
	}
	//if err := func(request *http.Request) error {
	//	if request.Method == http.MethodPost {
	//		if err := upload(request); err != nil {
	//			return err
	//		}
	//		//_, _ = fmt.Fprintf(writer, `<script>window.location.href = "/problem/"</script>`)
	//	}
	//	return nil
	//}(request); err != nil {
	//	alert(writer, err)
	//}
}

func upload(request *http.Request) (*multipart.FileHeader, error) {
	file, handler, err := request.FormFile("mesh_file")
	if err != nil {
		return nil, err
	}
	defer func() {
		err = file.Close()
	}()
	//fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	//fmt.Printf("File Size: %+v\n", handler.Size)
	//fmt.Printf("MIME Header: %+v\n", handler.Header)
	fileData := make([]byte, handler.Size)
	l, err := file.Read(fileData)
	if int64(l) != handler.Size || err != nil {
		return nil, fmt.Errorf("unable to download file")
	}
	meshName := "downloads/" + handler.Filename
	err = os.WriteFile(meshName, fileData, 0644)
	if err != nil {
		return nil, fmt.Errorf("unable to download file")
	}
	return handler, nil
}

func meshInfoPageHandler(writer http.ResponseWriter, request *http.Request) {
	var (
		handler *multipart.FileHeader
		err     error
	)
	if request.URL.Path != "/info/" {
		http.NotFound(writer, request)
		return
	}
	if err = request.ParseForm(); err != nil {
		log.Fatal("500 Internal Server Error: ", err)
	}
	if err = func(request *http.Request) error {
		if request.Method == http.MethodPost {
			if handler, err = upload(request); err != nil {
				return err
			}
			//_, _ = fmt.Fprintf(writer, `<script>window.location.href = "/problem/"</script>`)
		}
		return nil
	}(request); err != nil {
		alert(writer, err)
	} else {
		if err = tmpl.ExecuteTemplate(writer, "info.html", handler); err != nil {
			log.Fatal("500 Internal Server Error: ", err)
		}
	}
}
