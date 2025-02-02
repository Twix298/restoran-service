package main

import (
	"fmt"
	"log"
	"main/internal/app"
	"main/internal/db"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
)

func main() {
	indexName := "places"
	client, err := elasticsearch.NewDefaultClient()

	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	_, err = client.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	respCode, err := client.Indices.Exists([]string{"places"})
	if err != nil {
		log.Fatalf("Error checking index: %v", err)
	}

	if respCode.StatusCode != http.StatusOK {
		log.Println("creating index..")
		places, err := app.ParseData()
		if err != nil {
			log.Fatalf("Error parsing data: %v", err)
		}
		err = app.SaveToDb(indexName, places, client)
		if err != nil {
			log.Fatalf("Error saving data to DB: %v", err)
		}

		log.Println("Data successfully saved to DB!")
		fmt.Printf("Finish. All data save in database")
	} else {
		log.Println("Index already exists, skipping data loading.")
	}
	store := db.NewStore(indexName, client)
	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", NewRouter(&store)))
	// http.Handle("/", NewRouter(&store))
	// log.Println("Server is running on port 8080")
	// log.Fatal(http.ListenAndServe(":8080", nil))

}
