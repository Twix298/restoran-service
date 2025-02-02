package types

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

type RequestStruct struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source Plase `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
