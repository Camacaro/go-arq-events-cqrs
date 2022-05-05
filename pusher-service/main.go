package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Camacaro/cqrs/events"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	NatsAddress string `envconfig:"NATS_ADDRESS"`
}

func main() {
	var cfg Config
	// No le pasamos ningun prefijo ya que van a estar disponible en el entorno de docker
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatalf("Failed to process config: %v", err)
	}

	// Creamos la conexion a websockets
	hub := NewHub()

	// Conectamos a Nats
	natsAddress := fmt.Sprintf("nats://%s", cfg.NatsAddress)
	n, err := events.NewNatsEventStore(natsAddress)
	if err != nil {
		log.Fatalf("Failed to create nats event store: %v", err)
	}

	err = n.OnCreatedFeed(func(m events.CreateFeedMessage) {
		feedMessage := newCreatedFeedMessage(m.ID, m.Title, m.Description, m.CreatedAt)
		// Enviamos el mensaje a todos los clientes conectados por websockets
		hub.Broadcast(feedMessage, nil)
	})

	if err != nil {
		log.Fatalf("Failed to subscribe to created feed event: %v", err)
	}

	events.SetEventStore(n)
	defer events.Close()

	// corremos el servidor de websockets
	go hub.Run()

	// Iniciamos el servidor
	http.HandleFunc("/ws", hub.HandleWebSocket)
	err = http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
