#!/bin/sh

# todo
# docker build
# docker run db
docker network create orders-net


# Build application image 
docker build -t ordersystem .

# run db container
docker container rm -f pg 
docker run -d --name pg \
  --network orders-net \
    -p 5432:5432 \
  --env-file ./debug.env \
  -e PGDATA=/var/lib/postgresql/18/docker \
  --mount source=pg18data,target=/var/lib/postgresql/18/docker \
  postgres:18

# run application container 
docker container rm -f ordersystem
docker run -d --name ordersystem \
    --network orders-net \
    --env-file ./debug.env \
    -p 3000:3000 \
    ordersystem
