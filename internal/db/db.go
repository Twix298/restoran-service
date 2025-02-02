package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"main/types"

	"github.com/elastic/go-elasticsearch/v8"
)

type store struct {
	indexName string
	client    *elasticsearch.Client
}

type Store interface {
	// returns a list of items, a total number of hits and (or) an error in case of one
	GetPlaces(limit int, offset int) ([]types.Plase, int, error)
	GetPlacesByLocation(lat, lon float64) ([]types.Plase, int, error)
}

func NewStore(indexName string, client *elasticsearch.Client) Store {
	return &store{indexName: indexName, client: client}
}

func (s *store) GetPlaces(limit int, offset int) ([]types.Plase, int, error) {
	fmt.Printf("limit = %d, offset = %d\n", limit, offset)
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

	res, err := s.client.Search(
		s.client.Search.WithContext(context.Background()),
		s.client.Search.WithIndex("places"),
		s.client.Search.WithBody(&buf),
		s.client.Search.WithPretty(),
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
	var places []types.Plase
	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	for _, hit := range hits {
		source, exist := hit.(map[string]interface{})["_source"]
		if exist {
			places = append(places, types.Plase{
				Id:      uint64(source.(map[string]interface{})["id"].(float64)),
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

func (s *store) GetPlacesByLocation(lat, lon float64) ([]types.Plase, int, error) {
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
	var result types.RequestStruct
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, 0, err
	}
	var places []types.Plase
	for _, value := range result.Hits.Hits {
		places = append(places, value.Source)
	}
	return places, result.Hits.Total.Value, nil
}
