FROM golang:1.23-alpine

RUN apk add --no-cache git curl && \
      go install github.com/air-verse/air@v1.61.7

WORKDIR /app
COPY go.mod .

RUN go mod download

COPY . .

COPY .air.toml .

CMD ["air"]