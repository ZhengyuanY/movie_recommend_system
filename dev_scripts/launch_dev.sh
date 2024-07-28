#!/bin/bash

SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

# clean up used network and containers
docker container stop elasticsearch
docker container stop movie-rec-dev
docker network rm movie-rec-net

# setup new network and containers
docker network create movie-rec-net

docker run -d --rm --name elasticsearch --network=movie-rec-net -v "$SCRIPTPATH/../reverse_index/data":/usr/share/elasticsearch/data -p 9200:9200 -e "discovery.type=single-node" elasticsearch:7.17.22

docker run -p 8080:8080 -v "$SCRIPTPATH/../autocomplete/":/app -it --rm --name=movie-rec-dev --network=movie-rec-net movie-rec:dev
