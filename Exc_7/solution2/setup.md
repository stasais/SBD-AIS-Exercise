# Docker Swarm Setup Commands

## 1. Init Swarm (Manager - Linux)
```bash
docker swarm init --advertise-addr 192.168.1.64
```

## 2. Join Swarm (Worker - Mac)
```bash
docker swarm join --token SWMTKN-1-4ssd0kvmf9ertxqubfmj43xz2vf7wqqj8iiwluu3175h1qutzl-bz5nep5drbrcamjiqvbnl6etq 192.168.1.64:2377
```

## 3. Verify Nodes
```bash
docker node ls
docker node ps
```

## 4. Deploy Stack
```bash
docker stack deploy --compose-file docker-compose.yml sbd
```

## 5. Check Services
```bash
docker service ls
```

## 6. Remove Stack
```bash
docker stack rm sbd
```