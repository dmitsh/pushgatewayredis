# PushGatewayRedis
Prometheus Pushgateway based on Redis.

### Rationale
[By design](https://prometheus.io/docs/practices/pushing/), Prometheus Pushgateway does not forget series pushed to it. However, sometimes that would be a desirable feature.
This PushGatewayRedis uses Redis to store the metrics, and leverages key-value expiration feature to remove stale metrics.
It supports Redis Server, Redis Cluster, and Redis Sentinel topologies.

### Overview
PushGatewayRedis connects to Redis instance and exposes two HTTP endpoints. One is used for metrics ingestion, the other acts as a Prometheus metrics endpoint.
To build PushGatewayRedis, execute
```sh
$ make build
```
Here is the usage:
```sh
$ pushgateway -h
usage: pushgateway [<flags>]

The Prometheus Redis Pushgateway

Flags:
  -h, --help                     Show context-sensitive help (also try --help-long and --help-man).
      --config.file=CONFIG.FILE  Prometheus configuration file path.
  -p, --port=9753                Service port.
      --tls.enabled              Enable TLS.
      --tls.key=TLS.KEY          Path to the server key.
      --tls.cert=TLS.CERT        Path to the server certificate.
  -m, --metrics.path="/metrics"  Metrics path.
  -i, --ingest.path="/ingest"    Ingest path.
      --redis.endpoint=":6379"   Redis endpoint(s).
      --redis.expiration=5m      Redis key/value expiration.
      --log.level=info           Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt        Output format of log messages. One of: [logfmt, json]
```
For more detailed settings use [config file](./examples/pushgw.yml) defined with this [schema](./pkg/config/config.go)

### End-to-end example
Start Redis.
```sh
$ make start-redis
```
Start PushGatewayRedis.
```sh
$ ./pushgateway
```
Start Prometheus server.
```sh
make start-prom
```
Ingest metrics from [sample data file](./examples/data.txt)
```sh
make ingest
```
Browse Prometheus UI at localhost:9090 and query metrics specified in the sample data file.
