package main

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/server"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	nodeId = "test-node"
)

var (
	logger          *zap.Logger
	snapshotCache   cache.SnapshotCache
	snapshotVersion =  uint64(1)
)

func init() {
	var err error
	logConfig := zap.NewDevelopmentConfig()
	logger, err = logConfig.Build()
	snapshotCache = cache.NewSnapshotCache(true, nodeHasher{}, logger.Sugar())
	if err != nil {
		panic(err)
	}
}

func main() {
	logger.Info("starting control-plane")

	// This wait seems necessary to trigger envoy attempting to connect before the control-plane is ready
	time.Sleep(15*time.Second)

	// Create the initial snapshot
	err := updateSnapShot()
	if err != nil {
		panic(err)
	}

	// Create an ADS server on port 10000
	adsSvr := server.NewServer(snapshotCache, nil)
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_zap.StreamServerInterceptor(logger))),
	)
	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, adsSvr)
	lis, err := net.Listen("tcp", ":10000")
	if err != nil {
		panic(err)
	}

	// Add an update function to trigger a new snapshot to be created and stored
	m := http.NewServeMux()
	m.HandleFunc("/update", handleUpdate)

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			<- ticker.C
			err := updateSnapShot()
			if err != nil {
				logger.Error("Error while updating snapshot", zap.Error(err))
				ticker.Stop()
				return
			}
		}
	}()

	// Start the http and grpc server
	go func() {
		logger.Fatal("grpc server returned error", zap.Error(grpcServer.Serve(lis)))
	}()
	logger.Info("grpc server started on port 10000")
	logger.Fatal("http server returned error", zap.Error(http.ListenAndServe(":9090", m)))
}

// http handler to trigger a new snapshot to be set in the snapshot cache
func handleUpdate(rw http.ResponseWriter, req *http.Request) {
	err := updateSnapShot()
	if err != nil {
		logger.Error("error updating", zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
	}
	rw.Write([]byte("snapshot updated"))
	rw.WriteHeader(http.StatusOK)
}

func updateSnapShot() error {
	nextVersion := atomic.AddUint64(&snapshotVersion, 1)
	nextVersionStr := strconv.FormatUint(nextVersion, 10)
	adsSnapshot := cache.NewSnapshot(
		nextVersionStr,
		endpoints(int(nextVersion)),
		clusters(int(nextVersion)),
		routes(),
		listeners())

	logger.Info("snapshot updated", zap.Uint64("version", nextVersion))
	err := snapshotCache.SetSnapshot(nodeId, adsSnapshot)
	if err != nil {
		return err
	}
	return nil
}

type nodeHasher struct{}

func (nodeHasher) ID(node *core.Node) string {
	return nodeId
}

type Callback struct{}

func (cb Callback) OnStreamOpen(context.Context, int64, string) error                   { return nil }
func (cb Callback) OnStreamClosed(int64)                                                {}
func (cb Callback) OnStreamRequest(int64, *v2.DiscoveryRequest)                         {}
func (cb Callback) OnStreamResponse(int64, *v2.DiscoveryRequest, *v2.DiscoveryResponse) {}
func (cb Callback) OnFetchRequest(context.Context, *v2.DiscoveryRequest) error          { return nil }
func (cb Callback) OnFetchResponse(*v2.DiscoveryRequest, *v2.DiscoveryResponse)         {}
