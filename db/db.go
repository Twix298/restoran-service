package db

import (
	"main/types"
)

type Store interface {
	// returns a list of items, a total number of hits and (or) an error in case of one
	GetPlaces(limit int, offset int) ([]types.Plase, int, error)
}
