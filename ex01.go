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

func indexHandler(w http.ResponseWriter, r *http.Request) {
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

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Error get wd: ", err)
	}
	fmt.Println(wd + "/template/index.html")
	var tpl = template.Must(template.ParseFiles(wd + "/template/index.html"))
	htmlStruct := struct {
		Places []types.Plase
		Total  int
		Prev   int
		Next   int
		Last   int
	}{
		Places: values,
		Total:  total,
		Prev:   page - 1,
		Next:   page + 1,
		Last:   (total / limit) - 1,
	}
	if err := tpl.Execute(w, htmlStruct); err != nil {
		fmt.Println("error code")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	port := 8888
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	http.ListenAndServe(":"+strconv.Itoa(port), mux)

}
