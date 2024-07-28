#!/bin/bash

# Define the Elasticsearch URL
ES_URL="localhost:9200"
INDEX_NAME="movies"
SCRIPT_PATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
DEV_DATA="dev_data"

# Function to check if index exists
check_index() {
  response=$(curl -s -o /dev/null -w "%{response_code}" -I "$ES_URL/$INDEX_NAME")
  if [ "$response" -eq 200 ]; then
    echo "Index $INDEX_NAME exists."
  else
    echo "Index $INDEX_NAME does not exist. Creating index..."
    create_index
  fi
}

# Function to create index with mappings
create_index() {
  curl -s -H "Content-Type: application/json" -XPUT "$ES_URL/$INDEX_NAME" -d '{
    "mappings": {
      "properties": {
        "title": { "type": "text" },
        "genre": { "type": "text" },
        "year": { "type": "integer" },
        "director": { "type": "text" },
        "cast": { "type": "text" }
      }
    }
  }'
  echo "Index $INDEX_NAME created."
}

# Function to bulk upload data
bulk_upload() {
  curl -s -H "Content-Type: application/x-ndjson" -XPOST "$ES_URL/_bulk" --data-binary "@$SCRIPT_PATH/$DEV_DATA"
  echo "Data uploaded to $INDEX_NAME."
}

# Main script execution
check_index
bulk_upload

