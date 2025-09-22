FROM golang:1.25-alpine

WORKDIR /app

RUN go install github.com/jackc/tern@2.3.3

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./bin/main ./cmd/api