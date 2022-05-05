package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Camacaro/cqrs/events"
	"github.com/Camacaro/cqrs/models"
	"github.com/Camacaro/cqrs/repository"
	"github.com/Camacaro/cqrs/search"
)

/*
	Encargada de manejar cuando se resiva un evento de tipo
	created_feed
*/
func onCreatedFeed(m events.CreateFeedMessage) {
	feed := models.Feed{
		ID:          m.ID,
		Title:       m.Title,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
	}

	// Recibimos un feed nuevo ahora lo vamos a indexar
	if err := search.IndexFeed(context.Background(), &feed); err != nil {
		log.Fatalf("Failed to index feed: %v", err)
	}
}

func listFeedHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error
	feeds, err := repository.ListFeeds(ctx)
	if err != nil {
		log.Printf("Failed to list feeds: %v", err)
		http.Error(w, "Failed to list feeds", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(feeds); err != nil {
		log.Printf("Failed to encode feeds: %v", err)
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error

	query := r.URL.Query().Get("q")
	if len(query) == 0 {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	feeds, err := search.SearchFeeds(ctx, query)
	if err != nil {
		log.Printf("Failed to search feeds: %v", err)
		http.Error(w, "Failed to search feeds", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(feeds); err != nil {
		log.Printf("Failed to encode feeds: %v", err)
	}
}
