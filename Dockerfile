FROM golang:1.22-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY cmd/ ./cmd/

RUN go vet ./...

CMD ["go", "run", "./cmd/producer/main.go"]
