package main

import (
	"net/http"

	"main/internal/app"
	"main/internal/db"

	"github.com/go-chi/chi/v5"
)

func NewRouter(store *db.Store) http.Handler {
	r := chi.NewRouter()
	handler := app.NewHandler(store)
	r.Get("/", handler.IndexHandler)
	r.Get("/api/places", handler.IndexJSONHandler)

	r.Group(func(r chi.Router) {
		r.Get("/api/get_token", handler.Login)
	})

	r.Group(func(r chi.Router) {
		r.Use(handler.JwtMiddleware)
		r.Get("/api/recommend", handler.Search)

	})
	return r
}
