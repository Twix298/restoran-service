package app

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"main/types"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func ParseData() ([]types.Plase, error) {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
	}

	execDir := filepath.Dir(execPath) // Директория, где лежит исполняемый файл
	filePath := filepath.Join(execDir, "..", "materials", "data.csv")

	log.Println("Opening file:", filePath) // Для диагностики

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error open file: %w", err)
	}
	var places []types.Plase
	reader := csv.NewReader(file)
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	for _, record := range records {
		if record[1] == "Name" {
			continue
		}
		longitude, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			return nil, fmt.Errorf("Error parsing longitude: %w", err)
		}

		latitude, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing latitude: %w", err)
		}

		convId, _ := strconv.ParseUint(record[0], 10, 64)
		place := types.Plase{
			Id:       convId + 1,
			Name:     record[1],
			Address:  record[2],
			Phone:    record[3],
			Location: types.Location{latitude, longitude},
		}
		places = append(places, place)
	}
	log.Println("Finish parse data.csv")
	return places, nil
}

func SaveToDb(indexName string, places []types.Plase, client *elasticsearch.Client) error {
	for i, place := range places {

		data, err := json.Marshal(place)
		if err != nil {
			return fmt.Errorf("Error marshalling data: %v", err)
		}

		req := esapi.IndexRequest{
			Index:      indexName,
			DocumentID: strconv.Itoa(i + 1),
			Body:       bytes.NewReader(data),
			Refresh:    "true",
		}

		res, err := req.Do(context.Background(), client)
		if err != nil {
			return fmt.Errorf("Error saving to Elasticsearch: %v", err)
		}
		log.Println("save to database ", (i + 1))
		defer res.Body.Close()

		if res.IsError() {
			return fmt.Errorf("Error response from Elasticsearch: %v", res)
		}
	}
	return nil
}
