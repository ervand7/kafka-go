#!/bin/bash

# shellcheck disable=SC2046
docker stop $(docker ps -a -q) && docker rm $(docker ps -a -q)

docker system prune -f

docker compose up -d

docker exec -it kafka kafka-topics.sh \
  --create \
  --topic orders \
  --partitions 3 \
  --replication-factor 1 \
  --if-not-exists \
  --bootstrap-server kafka:9092

docker exec -it kafka kafka-topics.sh \
  --create \
  --topic orders-taxed \
  --partitions 3 \
  --replication-factor 1 \
  --bootstrap-server kafka:9092

docker exec -it kafka kafka-topics.sh \
  --create \
  --topic orders-dlq \
  --partitions 3 \
  --replication-factor 1 \
  --bootstrap-server kafka:9092


docker compose up -d --scale consumer=2