
build: pushgateway
	go build ./cmd/pushgateway/

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

.PHONY: ingest metrics start-redis stop-redis start-prom stop-prom
