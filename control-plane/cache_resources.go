package main

import (
	"fmt"
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	"math"
	"time"
)

const (
	clusterName = "test-cluster"
)

func endpoints(n int) []cache.Resource {
	lbEndpoints := []*envoy_api_v2_endpoint.LbEndpoint{lbEndpointFromPort(8080)}

	var clas []cache.Resource
	for i := 0; i < n; i++ {
		clusterLoadAssignment := &v2.ClusterLoadAssignment{
			ClusterName: fmt.Sprintf("%s-%d", clusterName, i),
			Endpoints: []*envoy_api_v2_endpoint.LocalityLbEndpoints{
				{
					Locality: &envoy_api_v2_core.Locality{
						Region: "test-locality",
					},
					LbEndpoints: lbEndpoints,
				},
			},
			Policy: &v2.ClusterLoadAssignment_Policy{
				OverprovisioningFactor: &wrappers.UInt32Value{
					Value: math.MaxUint32,
				},
			},
		}
		clas = append(clas, clusterLoadAssignment)
	}
	return clas
}

func lbEndpointFromPort(port int) *envoy_api_v2_endpoint.LbEndpoint {
	return &envoy_api_v2_endpoint.LbEndpoint{
		HostIdentifier: &envoy_api_v2_endpoint.LbEndpoint_Endpoint{
			Endpoint: &envoy_api_v2_endpoint.Endpoint{
				Address: &envoy_api_v2_core.Address{
					Address: &envoy_api_v2_core.Address_SocketAddress{
						SocketAddress: &envoy_api_v2_core.SocketAddress{
							Address: "0.0.0.0",
							PortSpecifier: &envoy_api_v2_core.SocketAddress_PortValue{
								PortValue: uint32(port),
							},
						},
					},
				},
			},
		},
		HealthStatus: envoy_api_v2_core.HealthStatus_HEALTHY,
	}
}

func clusters(n int) []cache.Resource {
	ctPb := duration.Duration{
		Seconds:              int64((time.Duration(1) * time.Second).Seconds()),
		Nanos:                0,
	}
	var clusters []cache.Resource

	for i := 0; i < n; i++ {
		clusters = append(clusters,
			&v2.Cluster{
				Name: fmt.Sprintf("%s-%d", clusterName, i),
				ClusterDiscoveryType: &v2.Cluster_Type{
					Type: v2.Cluster_EDS,
				},
				ConnectTimeout: &ctPb,
				EdsClusterConfig: &v2.Cluster_EdsClusterConfig{
					EdsConfig: &envoy_api_v2_core.ConfigSource{
						ConfigSourceSpecifier: &envoy_api_v2_core.ConfigSource_Ads{
							Ads: &envoy_api_v2_core.AggregatedConfigSource{},
						},
					},
				},
			},
		)
	}

	return clusters
}

func routes() []cache.Resource {
	routeEntry := &envoy_api_v2_route.Route{
		Match: &envoy_api_v2_route.RouteMatch{
			PathSpecifier: &envoy_api_v2_route.RouteMatch_Prefix{
				Prefix: "/",
			},
		},
		Action: &envoy_api_v2_route.Route_Route{
			Route: &envoy_api_v2_route.RouteAction{
				ClusterSpecifier: &envoy_api_v2_route.RouteAction_ClusterHeader{
					ClusterHeader: "x-envoy-cluster-name",
				},
			},
		},
	}

	routeConfig := &v2.RouteConfiguration{
		Name: "egress_route",
		VirtualHosts: []*envoy_api_v2_route.VirtualHost{
			&envoy_api_v2_route.VirtualHost{
				Name:    "test_virtual_host",
				Domains: []string{"*"},
				Routes:  []*envoy_api_v2_route.Route{routeEntry},
			},
		},
	}
	return []cache.Resource{routeConfig}
}

func listeners() []cache.Resource {
	return []cache.Resource{}
}
