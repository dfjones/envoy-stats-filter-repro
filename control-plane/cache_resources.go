package main

import (
	"math"
	"time"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	protobuf_types "github.com/gogo/protobuf/types"
)

const (
	clusterName = "test-cluster"
)

func endpoints() []cache.Resource {
	lbEndpoints := []endpoint.LbEndpoint{lbEndpointFromPort(8080)}

	clusterLoadAssignment := &v2.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []endpoint.LocalityLbEndpoints{
			endpoint.LocalityLbEndpoints{
				Locality: &core.Locality{
					Region: "test-locality",
				},
				LbEndpoints: lbEndpoints,
			},
		},
		Policy: &v2.ClusterLoadAssignment_Policy{
			OverprovisioningFactor: &protobuf_types.UInt32Value{
				Value: math.MaxUint32,
			},
		},
	}
	return []cache.Resource{clusterLoadAssignment}
}

func lbEndpointFromPort(port int) endpoint.LbEndpoint {
	return endpoint.LbEndpoint{
		Endpoint: &endpoint.Endpoint{
			Address: &core.Address{
				Address: &core.Address_SocketAddress{
					SocketAddress: &core.SocketAddress{
						Address: "0.0.0.0",
						PortSpecifier: &core.SocketAddress_PortValue{
							PortValue: uint32(port),
						},
					},
				},
			},
		},
		HealthStatus: core.HealthStatus_HEALTHY,
	}
}

func clusters() []cache.Resource {
	return []cache.Resource{
		&v2.Cluster{
			Name:           clusterName,
			Type:           v2.Cluster_EDS,
			ConnectTimeout: time.Duration(1) * time.Second,
			EdsClusterConfig: &v2.Cluster_EdsClusterConfig{
				EdsConfig: &core.ConfigSource{
					ConfigSourceSpecifier: &core.ConfigSource_Ads{
						Ads: &core.AggregatedConfigSource{},
					},
				},
			},
		},
	}
}

func routes() []cache.Resource {
	routeEntry := route.Route{
		Match: route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: "/",
			},
		},
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_ClusterHeader{
					ClusterHeader: "x-envoy-cluster-name",
				},
			},
		},
	}

	routeConfig := &v2.RouteConfiguration{
		Name: "egress_route",
		VirtualHosts: []route.VirtualHost{
			route.VirtualHost{
				Name:    "test_virtual_host",
				Domains: []string{"*"},
				Routes:  []route.Route{routeEntry},
			},
		},
	}
	return []cache.Resource{routeConfig}
}

func listeners() []cache.Resource {
	return []cache.Resource{}
}
