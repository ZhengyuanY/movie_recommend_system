# syntax=docker/dockerfile:1

##
## Build the application from source
##

FROM golang:1.22 AS movie-rec

WORKDIR /app

##
## Run the tests in the container
##

FROM movie-rec AS test-movie-rec
RUN go test -v ./...

