package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/Camacaro/cqrs/models"
	elastic "github.com/elastic/go-elasticsearch/v7"
)

type ElasticSearchRepository struct {
	client *elastic.Client
}

func NewElasticSearchRepository(url string) (*ElasticSearchRepository, error) {
	// Crear la conexion a elasticsearch
	client, err := elastic.NewClient(elastic.Config{
		Addresses: []string{url},
	})

	if err != nil {
		return nil, err
	}

	return &ElasticSearchRepository{
		client: client,
	}, nil
}

func (e *ElasticSearchRepository) Close() {
	// Cerrar la conexion a elasticsearch
	// Pero elasticsearch no tiene una funcion para cerrar la conexion
}

func (e *ElasticSearchRepository) IndexFeed(ctx context.Context, feed *models.Feed) error {
	// Nos da la representacion en bytes
	body, _ := json.Marshal(feed)
	// Indexar el feed
	_, err := e.client.Index(
		"feeds",                                // index name
		bytes.NewReader(body),                  // procesar el body y darselo de una forma mas legible a elasticsearch
		e.client.Index.WithDocumentID(feed.ID), // document id
		e.client.Index.WithContext(ctx),        // context, esto nos ayudara hacer debug de la aplicacion en caso de que algo salga mal
		e.client.Index.WithRefresh("wait_for"), // refresh, para que refresque los diferentes indices
	)

	return err
}

func (e *ElasticSearchRepository) SearchFeeds(ctx context.Context, query string) (results []models.Feed, err error) {
	var buf bytes.Buffer
	/*
		JSON
		{
			query: "a", 1, true, [], {}
		}

		Por eso creamos un map de llaves de tipo string pero de valores de tipo interface{}
		osea any, esta es la manera de poder implementar un tipo de dato flexible
		en golang, como lo es un JSON
	*/

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":            query,                            // Valor
				"fields":           []string{"title", "description"}, // donde quiero buscar
				"fuziness":         3,                                // porcentaje de coincidencia, 3 significa que debe coincidir al menos 3 porcientos, por ejemplo gou -> go y pueda encontrar go
				"cutoff_frequency": 0.0001,                           // nos ayuda a decir cuantas veces debe coincidir para que sea considerado, para que devuleva los resultados mas relevantes
			}, // Este le indica a elasticsearch que queremos traer multiples elementos
		},
	}

	// Mandamos a codificar la query
	if err = json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, err
	}

	// Realizamos la busqueda
	res, err := e.client.Search(
		e.client.Search.WithContext(ctx),         // El contexto
		e.client.Search.WithIndex("feeds"),       // El index a buscar
		e.client.Search.WithBody(&buf),           // El body de la query
		e.client.Search.WithTrackTotalHits(true), // Nos ayuda a saber cuantos resultados hay
	)

	if err != nil {
		return nil, err
	}

	defer func() {
		// cerrar la conexion a elasticsearch
		if err := res.Body.Close(); err != nil {
			results = nil
		}
	}()

	// Error al hacer la consulta
	if res.IsError() {
		// devolvemos la representacion del error a nivel de string
		return nil, errors.New(res.String())
	}

	//  ==== Ahora hay que hacer un procesor de los resultados ===
	var eRes map[string]interface{}
	// Decodificar el body y lo guardamos en un map
	if err := json.NewDecoder(res.Body).Decode(&eRes); err != nil {
		return nil, err
	}

	/*
		map[string]interface{} -> es igual a un any, es decirle a golang que puede ser cualquier cosa
		un objeto por lo cual podemos hacer un cast a un map[string]interface{}
		y acceder al campo
	*/
	var feeds []models.Feed
	// hit es cada uno de los elementos que hizo match con la query
	for _, hit := range eRes["hits"].(map[string]interface{})["hits"].([]interface{}) {
		// hit es un map de llaves de tipo string pero de valores de tipo interface{}
		feed := models.Feed{}
		source := hit.(map[string]interface{})["_source"] // source es el contenido del hit, una propiedad de hit
		marshal, err := json.Marshal(source)              // marshal es la representacion en bytes de source, source -> bytes
		if err != nil {
			return nil, err
		}

		// unmarshal es la representacion en bytes de source, bytes -> source
		// Con esto podemos pasar los datos de source a una feed
		if err := json.Unmarshal(marshal, &feed); err == nil {
			// Si el proceso de unmarshal no da error, entonces podemos agregar el feed a la lista
			feeds = append(feeds, feed)
		}
	}

	return feeds, nil
}
