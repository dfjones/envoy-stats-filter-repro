## Envoy 5555 Repro

This repository contains a small repro for envoy issue 5555

### Running
Start with `docker-compose up -d --build`

Examine `localhost:9901/clusters`

curl `localhost:9090/update` to trigger the snapshot to be updated

Examine `localhost:9901/clusters`, notice clusters now contain ip addresses




