你好！
很冒昧用这样的方式来和你沟通，如有打扰请忽略我的提交哈。我是光年实验室（gnlab.com）的HR，在招Golang开发工程师，我们是一个技术型团队，技术氛围非常好。全职和兼职都可以，不过最好是全职，工作地点杭州。
我们公司是做流量增长的，Golang负责开发SAAS平台的应用，我们做的很多应用是全新的，工作非常有挑战也很有意思，是国内很多大厂的顾问。
如果有兴趣的话加我微信：13515810775  ，也可以访问 https://gnlab.com/，联系客服转发给HR。
## Envoy xDS deadlock repro

This repository contains a small reproduction of an Envoy issue where Envoy fails to subscribe to CDS updates from the control-plane when a `stats_matcher` `inclusion_list` is active.
See [envoy-config.yaml](https://github.com/dfjones/envoy-stats-filter-repro/blob/master/envoy-config.yaml#L14).

Here, the control-plane is based off of [envoyproxy/go-control-plane](https://github.com/envoyproxy/go-control-plane).

If an ADS control-plane connection is used and CDS fails, if a new Cluster is added, EDS will then fail.


### Variations

* [`master`](https://github.com/dfjones/envoy-stats-filter-repro/tree/master) branch: Envoy v1.12.2 and go-control-plane v0.8.6
* [`latest-control-plane`](https://github.com/dfjones/envoy-stats-filter-repro/tree/latest-control-plane) branch: Envoy v1.12.2 and go-control-plane v0.9.2
* [`envoy-1.13.0`](https://github.com/dfjones/envoy-stats-filter-repro/tree/envoy-1.13.0) branch: Envoy v1.13.0 and go-control-plane v0.9.2
* [`no-stats-filter`](https://github.com/dfjones/envoy-stats-filter-repro/tree/no-stats-filter) branch: stats inclusion list disabled, Envoy v1.13.0 and go-control-plane v0.9.2
* [`min-stats-repro`](https://github.com/dfjones/envoy-stats-filter-repro/tree/min-stats-repro) branch: shows the most minimal list of stats exclusion that causes the issue, Envoy v1.13.0 and go-control-plane v0.9.2
  * This branch shows that excluding just `- suffix: "warming_clusters"` reproduces the issue. 

### Running
Start with `docker-compose up -d --build && docker-compose logs -f control-plane`

This will start the system and tail the control-plane logs.

The control-plane sleeps for 15 seconds before binding its grpc port. 
This delay is necessary to reproduce the issue.

Example output:

```
control-plane_1  | 2020-01-29T22:37:33.417Z	INFO	go/main.go:43	starting control-plane
control-plane_1  | 2020-01-29T22:37:48.420Z	INFO	go/main.go:112	snapshot updated	{"version": 2}
control-plane_1  | 2020-01-29T22:37:48.420Z	INFO	go/main.go:86	grpc server started on port 10000
control-plane_1  | 2020-01-29T22:37:52.157Z	DEBUG	cache/simple.go:242	respond type.googleapis.com/envoy.api.v2.Cluster[] version "" with version "2"
control-plane_1  | 2020-01-29T22:37:52.157Z	DEBUG	cache/simple.go:242	respond type.googleapis.com/envoy.api.v2.RouteConfiguration[egress_route] version "" with version "2"
control-plane_1  | 2020-01-29T22:37:52.160Z	DEBUG	cache/simple.go:242	respond type.googleapis.com/envoy.api.v2.ClusterLoadAssignment[test-cluster-1 test-cluster-0] version "" with version "2"
control-plane_1  | 2020-01-29T22:37:52.165Z	DEBUG	cache/simple.go:195	open watch 1 for type.googleapis.com/envoy.api.v2.RouteConfiguration[egress_route] from nodeID "test-node", version "2"
control-plane_1  | 2020-01-29T22:37:52.166Z	DEBUG	cache/simple.go:195	open watch 2 for type.googleapis.com/envoy.api.v2.ClusterLoadAssignment[test-cluster-1 test-cluster-0] from nodeID "test-node", version "2"
control-plane_1  | 2020-01-29T22:37:53.421Z	INFO	go/main.go:112	snapshot updated	{"version": 3}
control-plane_1  | 2020-01-29T22:37:53.421Z	DEBUG	cache/simple.go:113	respond open watch 1[egress_route] with new version "3"
control-plane_1  | 2020-01-29T22:37:53.422Z	DEBUG	cache/simple.go:242	respond type.googleapis.com/envoy.api.v2.RouteConfiguration[egress_route] version "2" with version "3"
control-plane_1  | 2020-01-29T22:37:53.422Z	DEBUG	cache/simple.go:113	respond open watch 2[test-cluster-1 test-cluster-0] with new version "3"
control-plane_1  | 2020-01-29T22:37:53.422Z	DEBUG	cache/simple.go:236	ADS mode: not responding to request: "test-cluster-2" not listed
control-plane_1  | 2020-01-29T22:37:53.424Z	DEBUG	cache/simple.go:195	open watch 3 for type.googleapis.com/envoy.api.v2.RouteConfiguration[egress_route] from nodeID "test-node", version "3"
control-plane_1  | 2020-01-29T22:37:58.421Z	INFO	go/main.go:112	snapshot updated	{"version": 4}
control-plane_1  | 2020-01-29T22:37:58.421Z	DEBUG	cache/simple.go:113	respond open watch 3[egress_route] with new version "4"
control-plane_1  | 2020-01-29T22:37:58.421Z	DEBUG	cache/simple.go:242	respond type.googleapis.com/envoy.api.v2.RouteConfiguration[egress_route] version "3" with version "4"
control-plane_1  | 2020-01-29T22:37:58.423Z	DEBUG	cache/simple.go:195	open watch 4 for type.googleapis.com/envoy.api.v2.RouteConfiguration[egress_route] from nodeID "test-node", version "4"
control-plane_1  | 2020-01-29T22:38:03.425Z	INFO	go/main.go:112	snapshot updated	{"version": 5}
control-plane_1  | 2020-01-29T22:38:03.425Z	DEBUG	cache/simple.go:113	respond open watch 4[egress_route] with new version "5"
control-plane_1  | 2020-01-29T22:38:03.425Z	DEBUG	cache/simple.go:242	respond type.googleapis.com/envoy.api.v2.RouteConfiguration[egress_route] version "4" with version "5"
control-plane_1  | 2020-01-29T22:38:03.427Z	DEBUG	cache/simple.go:195	open watch 5 for type.googleapis.com/envoy.api.v2.RouteConfiguration[egress_route] from nodeID "test-node", version "5"
control-plane_1  | 2020-01-29T22:38:08.422Z	INFO	go/main.go:112	snapshot updated	{"version": 6}
control-plane_1  | 2020-01-29T22:38:08.422Z	DEBUG	cache/simple.go:113	respond open watch 5[egress_route] with new version "6"
control-plane_1  | 2020-01-29T22:38:08.422Z	DEBUG	cache/simple.go:242	respond type.googleapis.com/envoy.api.v2.RouteConfiguration[egress_route] version "5" with version "6"
control-plane_1  | 2020-01-29T22:38:08.424Z	DEBUG	cache/simple.go:195	open watch 6 for type.googleapis.com/envoy.api.v2.RouteConfiguration[egress_route] from nodeID "test-node", version "6"
```

Here we see that Envoy eventually connects and receives CDS and RDS responses at version 2.

Initial EDS responses are sent until there is an inconsistency between Envoy and the control-plane on the list of Clusters:
`control-plane_1  | 2020-01-29T22:37:53.422Z	DEBUG	cache/simple.go:236	ADS mode: not responding to request: "test-cluster-2" not listed`
See [go-control-plane/issues/119](https://github.com/envoyproxy/go-control-plane/issues/119) for background on this behavior.

In these logs we can see that no watch is created for `envoy.api.v2.Cluster[]` (CDS) after the initial update at version 2. 
While I have not determined the exact failure here, a likely reason is that Envoy has responded with an incorrect xDS version or nonce causing the control-plane to, correctly, ignore the message.

The xDS protocol will then continue in this state with no further CDS or EDS updates being sent to Envoy.




