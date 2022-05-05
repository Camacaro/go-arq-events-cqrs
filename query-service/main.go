package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Camacaro/cqrs/database"
	"github.com/Camacaro/cqrs/events"
	"github.com/Camacaro/cqrs/repository"
	"github.com/Camacaro/cqrs/search"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
)

// ESte sera nuestro Command de CQRS, Encargado de escribir en base de datos
type Config struct {
	PostgresDB           string `envconfig:"POSTGRES_DB"` // Mismo nombre que el docker compose
	PostgresUser         string `envconfig:"POSTGRES_USER"`
	PostgresPassword     string `envconfig:"POSTGRES_PASSWORD"`
	NatsAddress          string `envconfig:"NATS_ADDRESS"`
	ElasticsearchAddress string `envconfig:"ELASTICSEARCH_ADDRESS"`
}

func newRouter() (router *mux.Router) {
	router = mux.NewRouter()
	router.HandleFunc("/feeds", listFeedHandler).Methods(http.MethodGet)
	router.HandleFunc("/search", searchHandler).Methods(http.MethodGet)
	return
}

/*
	Q: CQRS
	Query Service
*/
func main() {
	var cfg Config
	// No le pasamos ningun prefijo ya que van a estar disponible en el entorno de docker
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatalf("Failed to process config: %v", err)
	}

	// Conectamos a los diferentes servicios

	// Conectamos a Postgres
	postgresAddress := fmt.Sprintf("postgres://%s:%s@postgres/%s?sslmode=disable", cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB)
	repo, err := database.NewPostgreRepository(postgresAddress)
	if err != nil {
		log.Fatalf("Failed to create postgres repository: %v", err)
	}
	repository.SetRepository(repo)

	// Conectamos a Elasticsearch
	elasticsearchAddress := fmt.Sprintf("http://%s", cfg.ElasticsearchAddress)
	es, err := search.NewElasticSearchRepository(elasticsearchAddress)
	if err != nil {
		log.Fatalf("Failed to create elasticsearch repository: %v", err)
	}
	search.SetRepository(es)
	defer search.Close()

	// Conectamos a Nats
	natsAddress := fmt.Sprintf("nats://%s", cfg.NatsAddress)
	n, err := events.NewNatsEventStore(natsAddress)
	if err != nil {
		log.Fatalf("Failed to create nats event store: %v", err)
	}
	/*
		Nuestro query se subscriba a un evento
		e indexea en elasticsearch
	*/
	err = n.OnCreatedFeed(onCreatedFeed)
	if err != nil {
		log.Fatalf("Failed to subscribe to created feed event: %v", err)
	}
	events.SetEventStore(n)
	defer events.Close()

	// Router
	router := newRouter()
	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
