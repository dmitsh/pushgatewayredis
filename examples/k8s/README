# Running pushgateway in Kubernetes

##### 1. Deploy Redis server
```sh
kubectl apply -f redis
```
##### 2. Deploy pushgateway
```sh
kubectl apply -f pushgw
```
##### Deploy a Prometheus target and start sending metrics to the pushgateway
```sh
kubectl apply -f input
```
##### Deploy Prometheus server and start ingesting metrics
```sh
kubectl apply -f prometheus
```
