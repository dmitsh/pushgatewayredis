DOCKER_IMAGE_VER=0.1

DOCKER_CONTAINER=pushgatewayredis:${DOCKER_IMAGE_VER}

build:
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' ./cmd/pushgateway/

ingest:
	curl -X POST -H "Content-Type: text/plain" --data-binary @./examples/data.txt localhost:9753/ingest

metrics:
	curl -X GET localhost:9753/metrics

start-redis:
	docker run -d -p 6379:6379 --name redis-srv redis redis-server

stop-redis:
	docker kill redis-srv; docker rm redis-srv

start-prom:
	docker run -d -p 9090:9090 --name prom-srv -v ${PWD}/examples/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus

stop-prom:
	docker kill prom-srv; docker rm prom-srv

docker-build:
	docker build -t ${DOCKER_CONTAINER} .

docker-push:
	docker tag ${DOCKER_CONTAINER} docker.io/dmitsh/${DOCKER_CONTAINER} && docker push docker.io/dmitsh/${DOCKER_CONTAINER}

.PHONY: build ingest metrics start-redis stop-redis start-prom stop-prom docker-build docker-push
