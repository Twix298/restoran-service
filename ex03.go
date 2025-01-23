package main

import (
	"encoding/json"
	"log"
	"main/types"
	"net/http"
	"strconv"
)

var data types.Plase

func showError(w *http.ResponseWriter) {
	// errorMessage := fmt.Sprintf("Invalid 'page' value: %s", pageQuery)
	http.Error(*w, "Bad request", http.StatusBadRequest)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	lat, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lon, err := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	if err != nil {
		showError(&w)
		return
	}
	places, _, err := data.GetPlacesByLocation(lat, lon)
	if err != nil {
		showError(&w)
	}
	if err != nil {
		log.Fatal("error while getting places: ", err)
	}

	jsonStruct := struct {
		Places []types.Plase `json:"places"`
	}{
		Places: places,
	}
	w.Header().Set("Content-Type", "application/json")
	jsonData, err := json.MarshalIndent(jsonStruct, "", "	")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(jsonData)
}

func main() {
	port := 8888
	mux := http.NewServeMux()
	mux.HandleFunc("/api/recommend", searchHandler)
	http.ListenAndServe(":"+strconv.Itoa(port), mux)
}
