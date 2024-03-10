package pkg

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
)

var (
	templates *template.Template
	query     string
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		ErrorHandler(w, http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		ErrorHandler(w, http.StatusMethodNotAllowed)
		return
	}

	templates, err := template.ParseGlob("./web/templates/*.html")
	if err != nil {
		log.Println(err)
		ErrorHandler(w, http.StatusInternalServerError)
		return
	}
	err = templates.ExecuteTemplate(w, "index.html", &ResponseData)
	if err != nil {
		log.Println(err)
		ErrorHandler(w, http.StatusInternalServerError)
		return
	}
}

func BandHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/band" {
		ErrorHandler(w, http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		ErrorHandler(w, http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")

	numID, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
		ErrorHandler(w, http.StatusInternalServerError)
		return
	}

	if numID > len(ResponseData.Band) {
		log.Println(err)
		ErrorHandler(w, http.StatusNotFound)
		return
	}

	band := ResponseData.Band[numID-1]

	band.Relations = RelationInfo.Index[numID-1].DatesLocations

	templates, err := template.ParseGlob("./web/templates/*.html")
	if err != nil {
		log.Println(err)
		ErrorHandler(w, http.StatusInternalServerError)
		return
	}

	err = templates.ExecuteTemplate(w, "band.html", &band)
	if err != nil {
		log.Println(err)
		ErrorHandler(w, http.StatusInternalServerError)
		return
	}
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/search" {
		ErrorHandler(w, http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		ErrorHandler(w, http.StatusMethodNotAllowed)
		return
	}

	query = r.URL.Query().Get("query")
	
	band, err := SearchRecords(ResponseData.Band, query)
	if err != nil {
		log.Println(err)
		NotFoundHandler(w, http.StatusNotFound)
		return
	}
	templates, err := template.ParseGlob("./web/templates/*.html")
	if err != nil {
		log.Println(err)
		ErrorHandler(w, http.StatusInternalServerError)
		return
	}
	err = templates.ExecuteTemplate(w, "search.html", &band)
	if err != nil {
		log.Println(err)
		ErrorHandler(w, http.StatusInternalServerError)
		return
	}
}

func ErrorHandler(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)

	data := struct {
		StatusMsg  string
		StatusCode int
	}{
		"Ooops. Error ",
		statusCode,
	}

	templates, err := template.ParseGlob("./web/templates/*.html")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = templates.ExecuteTemplate(w, "error.html", &data)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func NotFoundHandler(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
	data := struct {
		StatusMsg  string
		StatusCode int
	}{
		"We don't have information about this member or group yet :(",
		statusCode,
	}
	templates, err := template.ParseGlob("./web/templates/*.html")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = templates.ExecuteTemplate(w, "error.html", &data)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
