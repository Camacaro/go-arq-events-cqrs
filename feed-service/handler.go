package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Camacaro/cqrs/events"
	"github.com/Camacaro/cqrs/models"
	"github.com/Camacaro/cqrs/repository"
	"github.com/segmentio/ksuid"
)

type createFeedRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func createdFeedHandler(w http.ResponseWriter, r *http.Request) {
	var req createFeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdAt := time.Now().UTC()
	id, err := ksuid.NewRandom()
	if err != nil {
		log.Fatalf("failed to create id: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feed := models.Feed{
		ID:          id.String(),
		Title:       req.Title,
		Description: req.Description,
		CreatedAt:   createdAt,
	}

	// Save feed in database repository
	if err := repository.InsertFeed(r.Context(), &feed); err != nil {
		log.Fatalf("failed to insert feed: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Publish feed to event bus -> NATS
	if err := events.PublishCreatedFeed(r.Context(), &feed); err != nil {
		log.Fatalf("failed to publish feed created: %s", err)
	}

	// Return created feed
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(feed)
}
