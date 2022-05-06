# Curso de Go Avanzado: Arquitectura de Eventos y CQRS

Modulo 
```go mod init github.com/Camacaro/```

### Modulos Usados
* Postgresql -> ```go get github.com/lib/pq```
* NATS -> ```go get github.com/nats-io/nats.go```
* Elastisearch -> ```go get github.com/elastic/go-elasticsearch/v7```
* Server Route -> ```go get github.com/gorilla/mux```
* Manejar variables de entorno -> ```go get github.com/kelseyhightower/envconfig```
* Generar diferentes ID -> ```go get github.com/segmentio/ksuid```
* Websocket -> ```go get github.com/gorilla/websocket```

Indexacion de la data (ElasticSearch)
1. Hacer busqueda mas compleja 
2. El acceso o lectura a esas busqueda sean rapidas. 


Docker-composer
```docker-compose up -d --build```

docker compose V2 apple silicon
docker compose build
docker compose up