package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"main/types"
	"os"
	"strconv"
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func parse() []types.Plase {
	file, err := os.Open("./materials/data.csv")
	if err != nil {
		fmt.Println("error open:", err)
		return nil
	}
	var places []types.Plase
	reader := csv.NewReader(file)
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("error reading:", err)
		return nil
	}
	for _, record := range records {
		if record[1] == "Name" {
			continue
		}
		longitude, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			fmt.Println("Ошибка преобразования долготы:", err)
			return nil
		}

		latitude, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			fmt.Println("Ошибка преобразования широты:", err)
			return nil
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
	return places
}

func saveToDb(id int, value *types.Plase, client *elasticsearch.Client, wg *sync.WaitGroup) {
	defer wg.Done()
	data, err := json.Marshal(*value)

	if err != nil {
		fmt.Println("Error parsing data: ", err)
	}
	req := esapi.IndexRequest{
		Index:      "places",
		DocumentID: strconv.Itoa(id + 1),
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}
	res, err := req.Do(context.Background(), client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

}

func main() {
	places := parse()
	client, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
	res, err := client.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	respCode, err := client.Indices.Exists([]string{"places"})

	if respCode.StatusCode != 200 {
		log.Println("creating index..")
		res, err = client.Indices.Create("places")
	}

	log.Println(res)
	var wg sync.WaitGroup
	for i, val := range places {
		fmt.Println("i = ", i)
		wg.Add(1)
		saveToDb(i, &val, client, &wg)
		//go saveToDb(i, &val, client, &wg)
		if err != nil {
			log.Println("error save to database: ", err)
			break
		}
		wg.Wait()
	}
	fmt.Printf("Finish. All data save in database")
}
