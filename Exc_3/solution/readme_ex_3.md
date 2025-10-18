docker rm -f pg 2>/dev/null || true
docker run -d --name pg \
  --network host \
  --env-file ./debug.env \
  -e PGDATA=/var/lib/postgresql/18/docker \
  --mount source=pg18data,target=/var/lib/postgresql/18/docker \
  postgres:18


docker logs -f pg


PGPASSWORD=docker psql -h 127.0.0.1 -p 5432 -U docker -d order -c '\conninfo'