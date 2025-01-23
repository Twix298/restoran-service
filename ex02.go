package main

import (
	"fmt"
	"html/template"
	"log"
	"main/types"
	"net/http"
	"os"
	"strconv"
)

var data types.Plase

func showErrorPage(w *http.ResponseWriter, pageQuery string) {
	errorMessage := fmt.Sprintf("Invalid 'page' value: %s", pageQuery)
	http.Error(*w, errorMessage, http.StatusBadRequest)
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.RawQuery
	showErrorPage(&w, query)
}

func indexJSONHandler(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageQuery := r.URL.Query().Get("page")
	fmt.Println("pageQuery = ", pageQuery)
	if pageQuery != "" {
		page, _ = strconv.Atoi(pageQuery)
	} else {
		showErrorPage(&w, pageQuery)
		return
	}
	if page < 1 {
		showErrorPage(&w, pageQuery)
		return
	}
	limit := 10
	offset := ((page - 1) * limit)
	values, total, err := data.GetPlaces(limit, offset)
	if err != nil {
		log.Fatal("error while getting places: ", err)
	}

	jsonStruct := struct {
		Places []types.Plase `json:"places"`
		Total  int           `json:"total"`
		Prev   int           `json:"prev"`
		Next   int           `json:"next"`
		Last   int           `json:"last"`
	}{
		Places: values,
		Total:  total,
		Prev:   page - 1,
		Next:   page + 1,
		Last:   (total / limit) - 1,
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Error get wd: ", err)
	}
	tpl, err := template.ParseFiles(wd + "/template/index_ex02.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	if err := tpl.Execute(w, jsonStruct); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	port := 8888
	mux := http.NewServeMux()
	mux.HandleFunc("/api/places", indexJSONHandler)
	// mux.HandleFunc("/", errorHandler)
	http.ListenAndServe(":"+strconv.Itoa(port), mux)

}
