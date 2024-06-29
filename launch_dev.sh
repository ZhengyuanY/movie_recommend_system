#!/bin/bash

docker run -p 8080:8080 -v "$(pwd)":/app -it --rm --name movie-rec-dev movie-rec:dev
