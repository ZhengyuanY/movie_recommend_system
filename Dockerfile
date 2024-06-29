FROM golang:1.22

WORKDIR /app

EXPOSE 8080

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o ./docker-gs-ping

CMD ["./docker-gs-ping"]
