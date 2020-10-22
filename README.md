# PushGatewayRedis - Highly Available Prometheus Pushgateway based on Redis.

### Rationale
The main objective is to offer a highly available Prometheus Pushgateway, that will survive process or system crash.
The existing implementation of [Pushgateway](https://github.com/prometheus/pushgateway) does not persist metrics by default. The `--persistence.file` flag allows to specify a file in which the pushed metrics will be persisted, but this won't help in case of the node failure.
The proposed design uses Redis key-value cache to store and retrieve the metrics. It supports Redis Server, Redis Cluster, and Redis Sentinel topologies, thus, natively providing high-availability capabilities.

Another feature of the proposed pushgateway, also leveraged from Redis, is that it allows to automatically delete metrics after a specified duration.
[By design](https://prometheus.io/docs/practices/pushing/), the official Prometheus Pushgateway does not forget the series pushed to it. Instead, it allows to delete groups of metrics on-demand, which would require users to implement additional logic in their applications.
On the other hand, there are use cases where deletion of "stale" metrics would be preferable.
In any case, with the proposed pushgateway the auto-deletion of the metrics is configurable, and could be turned on and off.

*Note:* The proposed implementation lacks many features compared to the official Pushgateway. We would like to gather feedback from the open source community to understand whether this project has merit.

### Overview
PushGatewayRedis connects to Redis instance and exposes three HTTP endpoints:
- for metrics ingestion
- as a Prometheus metrics endpoint for collected metrics
- as a Prometheus metrics endpoint for Pushgateway telemetry

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
  -h, --help                        Show context-sensitive help (also try --help-long and --help-man).
      --config.file=CONFIG.FILE     Prometheus configuration file path.
  -p, --port=9753                   Service port.
      --tls.enabled                 Enable TLS.
      --tls.key=TLS.KEY             Path to the server key.
      --tls.cert=TLS.CERT           Path to the server certificate.
  -m, --metrics.path="/metrics"     Metrics URL path.
  -t, --telemetry.path="/telemetry" Telemetry URL path.
  -i, --ingest.path="/ingest"       Ingest URL path.
      --redis.endpoint=":6379"      Redis endpoint(s).
      --redis.expiration=5m         Redis key/value expiration.
      --log.level=info              Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt           Output format of log messages. One of: [logfmt, json]
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
