#!/bin/bash

SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
PRIVATEIP="$(ip -4 addr show enp42s0 | grep -oP '(?<=inet\s)\d+(\.\d+){3}')"
PUBLICIP="$(curl ifconfig.me)"
echo $PRIVATEIP
echo $PUBLICIP

# clean up used containers
docker container rm -f elasticsearch
docker container rm -f movie-rec-dev
docker container rm -f load-balancing

if [[ "$1" == "--stop" ]]; then
	exit 0
fi

# start new containers
docker run -d --rm -p 9200:9200 -v "$SCRIPTPATH/../reverse_index/data":/usr/share/elasticsearch/data -e "discovery.type=single-node" --name=elasticsearch elasticsearch:7.17.22

docker run -d --rm -p 8081:80 -p 8082:8080 -e AUTOCOMPLETE_IP=$PRIVATEIP -e REVERSE_INDEX_IP=$PRIVATEIP --name=load-balancing load-balancing 

docker run -d --rm -p 8080:8080 -e LOAD_BALANCER_IP=$PRIVATEIP -v "$SCRIPTPATH/../autocomplete/":/app --name=movie-rec-dev movie-rec:dev go run main.go #tail -f /dev/null

#docker exec -it movie-rec-dev "bash"
