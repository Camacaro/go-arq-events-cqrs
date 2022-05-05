# Feed Service

ARG GO_VERSION=1.16.6

FROM golang:${GO_VERSION}-alpine AS builder

# Esto es para la ruta de dependencias no va a usar ningun proxi que itulice la misma de cada paquete
RUN go env -w GOPROXY=direct

# Descargar paquetes para poder buildear
RUN apk add --no-cache git
RUN apk --no-cache add ca-certificates && update-ca-certificates

WORKDIR /src

COPY ./go.mod ./go.sum ./

RUN go mod download

# Copiar todas las carpetas
COPY events events
COPY database database
COPY repository repository
COPY search search
COPY models models
COPY feed-service feed-service
COPY query-service query-service

# Compilacion de todo el codigo
RUN go install ./...

FROM alpine:3.11
WORKDIR /usr/bin
COPY --from=builder /go/bin .
