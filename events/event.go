package events

import (
	"context"

	"github.com/Camacaro/cqrs/models"
)

/*
	Principio Inversion de dependencia
	para que nuestro codigo dependa de esta abstraccion
	y no de implementaciones concreta

	Para eventos, se puede usar NATS, RabbitMQ, Kafka, etc.
*/

type EventStore interface {
	Close()
	PublishCreatedFeed(ctx context.Context, feed *models.Feed) error            // Publicar cuando se ha creado un nuevo feed
	SubscribeCreatedFeed(ctx context.Context) (<-chan CreateFeedMessage, error) // subscriba para cuando un nuevo feed a sido creado y se pueda leer
	OnCreatedFeed(handler func(CreateFeedMessage)) error                        // Un callback que reaccione cuando un nuevo feed a sido creado
}

var eventStore EventStore

func Close() {
	eventStore.Close()
}

func PublishCreatedFeed(ctx context.Context, feed *models.Feed) error {
	return eventStore.PublishCreatedFeed(ctx, feed)
}

func SubscribeCreatedFeed(ctx context.Context) (<-chan CreateFeedMessage, error) {
	return eventStore.SubscribeCreatedFeed(ctx)
}

func OnCreatedFeed(handler func(CreateFeedMessage)) error {
	return eventStore.OnCreatedFeed(handler)
}
