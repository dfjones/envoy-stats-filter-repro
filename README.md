## Envoy xDS deadlock repro

This repository contains a small reproduction of an Envoy issue where Envoy fails to subscribe to CDS updates from the control-plane.

Here, the control-plane is based off of [envoyproxy/go-control-plane](https://github.com/envoyproxy/go-control-plane).

If an ADS control-plane connection is used and CDS fails, if a new Cluster is added, EDS will then fail.


### Variations

* `master` branch: Envoy v1.12.2 and go-control-plane v0.8.6
* `latest-control-plane` branch: Envoy v1.12.2 and go-control-plane v0.9.2
* `envoy-1.13.0` branch: Envoy v1.13.0 and go-control-plane v0.9.2

### Running
Start with `docker-compose up -d --build && docker-compose logs -f control-plane`

This will start the system and tail the control-plane logs.

The control-plane sleeps for 15 seconds before binding its grpc port. 
This delay is necessary to reproduce the issue.

Example output:

```
control-plane_1  | 2020-01-28T19:30:14.301Z	INFO	go/main.go:43	starting control-plane
control-plane_1  | 2020-01-28T19:30:29.270Z	INFO	go/main.go:111	snapshot updated	{"version": 2}
control-plane_1  | 2020-01-28T19:30:29.271Z	INFO	go/main.go:86	grpc server started on port 10000
control-plane_1  | 2020-01-28T19:30:34.273Z	INFO	go/main.go:111	snapshot updated	{"version": 3}
control-plane_1  | 2020-01-28T19:30:39.272Z	INFO	go/main.go:111	snapshot updated	{"version": 4}
control-plane_1  | 2020-01-28T19:30:43.905Z	INFO	cache/simple.go:243	respond type.googleapis.com/envoy.api.v2.Cluster[] version "" with version "4"
control-plane_1  | 2020-01-28T19:30:43.905Z	INFO	cache/simple.go:243	respond type.googleapis.com/envoy.api.v2.RouteConfiguration[egress_route] version "" with version "4"
control-plane_1  | 2020-01-28T19:30:43.953Z	INFO	cache/simple.go:237	ADS mode: not responding to request: "test-cluster" not listed
control-plane_1  | 2020-01-28T19:30:43.954Z	INFO	cache/simple.go:196	open watch 1 for type.googleapis.com/envoy.api.v2.RouteConfiguration[egress_route] from nodeID "test-node", version "4"
control-plane_1  | 2020-01-28T19:30:44.272Z	INFO	go/main.go:111	snapshot updated	{"version": 5}
control-plane_1  | 2020-01-28T19:30:44.273Z	INFO	cache/simple.go:114	respond open watch 1[egress_route] with new version "5"
control-plane_1  | 2020-01-28T19:30:44.273Z	INFO	cache/simple.go:243	respond type.googleapis.com/envoy.api.v2.RouteConfiguration[egress_route] version "4" with version "5"
control-plane_1  | 2020-01-28T19:30:44.276Z	INFO	cache/simple.go:196	open watch 2 for type.googleapis.com/envoy.api.v2.RouteConfiguration[egress_route] from nodeID "test-node", version "5"
control-plane_1  | 2020-01-28T19:30:49.237Z	INFO	go/main.go:111	snapshot updated	{"version": 6}
control-plane_1  | 2020-01-28T19:30:49.237Z	INFO	cache/simple.go:114	respond open watch 2[egress_route] with new version "6"
control-plane_1  | 2020-01-28T19:30:49.237Z	INFO	cache/simple.go:243	respond type.googleapis.com/envoy.api.v2.RouteConfiguration[egress_route] version "5" with version "6"
control-plane_1  | 2020-01-28T19:30:49.239Z	INFO	cache/simple.go:196	open watch 3 for type.googleapis.com/envoy.api.v2.RouteConfiguration[egress_route] from nodeID "test-node", version "6"
```

Here we see that Envoy eventually connects and receives CDS and RDS responses at version 4.

No EDS response is sent because there is not yet consistency between Envoy and the control-plane on the list of Clusters.
See [go-control-plane/issues/119](https://github.com/envoyproxy/go-control-plane/issues/119) for background on this behavior.

In these logs we can see that no watch is created for `envoy.api.v2.Cluster[]` (CDS) after the initial update at version 4. 
While I have not determined the exact failure here, a likely reason is that Envoy has responded with an incorrect xDS version or nonce causing the control-plane to, correctly, ignore the message.

The xDS protocol will then continue in this state with no further CDS or EDS updates being sent to Envoy.




