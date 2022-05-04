package events

import (
	"bytes"
	"context"
	"encoding/gob"

	"github.com/Camacaro/cqrs/models"
	"github.com/nats-io/nats.go"
)

type NatsEventStore struct {
	conn            *nats.Conn
	feedCreatedSub  *nats.Subscription     // Subscripcion para conectarse pcuan un nuevo feed a sido creado
	feedCreatedChan chan CreateFeedMessage // Canal para leer los mensajes de nuevo feed creado
}

func NewNatsEventStore(url string) (*NatsEventStore, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return &NatsEventStore{
		conn: conn,
	}, nil
}

func (n *NatsEventStore) Close() {
	if n.conn != nil {
		n.conn.Close()
	}

	if n.feedCreatedSub != nil {
		n.feedCreatedSub.Unsubscribe()
	}

	close(n.feedCreatedChan)
}

/*
	Funcion privada que nos va a permitir transmitir data
*/
func (n *NatsEventStore) encodeMessage(m Message) ([]byte, error) {
	b := bytes.Buffer{} // Buffer nuevo de tipo bytes
	// gob nos permite serializar una estructura
	// m sera encodiado en bytes
	err := gob.NewEncoder(&b).Encode(m)
	if err != nil {
		return nil, err
	}

	// Devuelve la representacion de bytes del mensaje
	return b.Bytes(), nil
}

// Publicar un mensaje, un nuevo feed
func (n *NatsEventStore) PublishCreatedFeed(ctx context.Context, feed *models.Feed) error {
	msg := CreateFeedMessage{
		ID:          feed.ID,
		Title:       feed.Title,
		Description: feed.Description,
		CreatedAt:   feed.CreatedAt,
	}

	// Encode el mensaje -> msg es aceptado por el encodeMessage ya que implementa una interfaz Message
	data, err := n.encodeMessage(msg)
	if err != nil {
		return err
	}

	// Publica el mensaje: (Evento, data)
	// Esto publicara a todos los servicios que esten subscrito a este evento
	return n.conn.Publish(msg.Type(), data)
}

/*
	Funcion privada que nos va a permitir recibir data
	decoding es el proceso de decodificacion
*/
func (n *NatsEventStore) decodeMessage(data []byte, m interface{}) error {
	// Creo un nuevo buffer, vacio
	b := bytes.Buffer{}
	// Le asigno el data que recibo
	b.Write(data)
	// Decodifico el buffer y le asigno una interfaz
	return gob.NewDecoder(&b).Decode(m)
}

// Esta funcion es un callback que se va a ejecutar cuando un nuevo feed a sido creado
func (n *NatsEventStore) OnCreatedFeed(handler func(CreateFeedMessage)) (err error) {
	msg := CreateFeedMessage{}

	n.feedCreatedSub, err = n.conn.Subscribe(msg.Type(), func(m *nats.Msg) {
		// Decodificar el mensaje
		n.decodeMessage(m.Data, &msg)
		// Ejecutar el callback
		handler(msg)
	})

	// Implicitamente, retorna la variable err
	return
}

/*
	Aqui nos subscribimos a un evento
	Nos subscribimos al evento created feed
*/
func (n *NatsEventStore) SubscribeCreatedFeed(ctx context.Context) (<-chan CreateFeedMessage, error) {
	m := CreateFeedMessage{}
	n.feedCreatedChan = make(chan CreateFeedMessage, 64)
	ch := make(chan *nats.Msg, 64)

	var err error
	n.feedCreatedSub, err = n.conn.ChanSubscribe(m.Type(), ch)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			// Este case se va activar cuando el ch reciba un mensaje
			case msg := <-ch:
				n.decodeMessage(msg.Data, &m) // Decodificar el mensaje
				n.feedCreatedChan <- m        // Enviar el mensaje al canal
			}
		}
	}()

	return (<-chan CreateFeedMessage)(n.feedCreatedChan), nil
}
