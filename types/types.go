package types

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Plase struct {
	Id       uint64   `json:"id"`
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Phone    string   `json:"phone"`
	Location Location `json:"location"`
}

type requestStruct struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source Plase `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (p *Plase) GetPlacesByLocation(lat, lon float64) ([]Plase, int, error) {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		return nil, 0, err
	}
	var buf bytes.Buffer
	query := map[string]interface{}{
		"track_total_hits": true,
		"sort": []map[string]interface{}{
			{
				"_geo_distance": map[string]interface{}{
					"location": map[string]float64{
						"lat": lat,
						"lon": lon,
					},
					"order":           "asc",
					"unit":            "km",
					"mode":            "min",
					"distance_type":   "arc",
					"ignore_unmapped": true,
				},
			},
		},
	}
	err = json.NewEncoder(&buf).Encode(query)
	if err != nil {
		return nil, 0, err
	}
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("places"),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()
	var result requestStruct
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, 0, err
	}
	var places []Plase
	for _, value := range result.Hits.Hits {
		places = append(places, value.Source)
	}
	return places, result.Hits.Total.Value, nil
}

func (p *Plase) GetPlaces(limit int, offset int) ([]Plase, int, error) {
	fmt.Printf("limit = %d, offset = %d\n", limit, offset)
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		return nil, 0, err
	}
	var buf bytes.Buffer
	query := map[string]interface{}{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []map[string]interface{}{
			{
				"id": map[string]interface{}{
					"order": "asc",
				},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("places"),
		es.Search.WithBody(&buf),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()
	if err != nil {
		fmt.Println("Error getting places", err)
		return nil, 0, err
	}
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	var places []Plase
	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	for _, hit := range hits {
		source, exist := hit.(map[string]interface{})["_source"]
		if exist {
			//location := source.(map[string]interface{})["location"].(map[string]interface{})
			// fmt.Println(source)
			// fmt.Printf("%v\n", source)
			// fmt.Println(uint64(source.(map[string]interface{})["id"].(float64)))
			places = append(places, Plase{
				Id:      uint64((source.(map[string]interface{}))["id"].(float64)),
				Name:    source.(map[string]interface{})["name"].(string),
				Address: source.(map[string]interface{})["address"].(string),
				Phone:   source.(map[string]interface{})["phone"].(string),
				//Lat: location["location"].(map[string]interface{})["lat"].(float64),
				//Lon: location["location"].(map[string]interface{})["lon"].(float64)},
			})
		}
	}

	totalHits := int(result["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))

	return places, totalHits, nil
}
