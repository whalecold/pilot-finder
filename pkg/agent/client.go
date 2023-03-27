package agent

import (
	"context"
	"fmt"
	"math"
	"time"

	v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/go-logr/logr"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/klog/v2"

	"github.com/whalecold/pilot-finder/pkg/constants"
)

const (
	defaultClientMaxReceiveMessageSize = math.MaxInt32
	defaultInitialConnWindowSize       = 1024 * 1024 // default gRPC InitialWindowSize
	defaultInitialWindowSize           = 1024 * 1024 // default gRPC ConnWindowSize
)

// Options the meta info of agent.
type Options struct {
	Type      string
	Address   string
	PodName   string
	Namespace string
	Suffix    string
}

func (opts *Options) NodeID() string {
	if opts.Suffix == "" {
		opts.Suffix = constants.NodeIDSuffix
	}
	return fmt.Sprintf("%s~%s~%s.%s~%s.%s", opts.Type, opts.Address, opts.PodName, opts.Namespace,
		opts.Namespace, opts.Suffix)
}

// Agent the interface of pilot xDS client agent.
type Agent interface {
	Send(req *discovery.DiscoveryRequest) error
	Close() error
	Connect(ctx context.Context) error
	Run(ctx context.Context, handler func(response *discovery.DiscoveryResponse))
	SetConnectLimit(limit *rate.Limiter)
}

type agent struct {
	pilotAddress string
	dialOptions  []grpc.DialOption

	stream discovery.AggregatedDiscoveryService_StreamAggregatedResourcesClient
	conn   *grpc.ClientConn

	node     *v3.Node
	opts     *Options
	respChan chan *discovery.DiscoveryResponse

	connRateLimit *rate.Limiter

	log logr.Logger
}

// Close ...
func (a *agent) Close() error {
	return a.conn.Close()
}

func (a *agent) SetConnectLimit(limit *rate.Limiter) {
	a.connRateLimit = limit
}

// New ...
func New(address string, opts *Options, log logr.Logger) Agent {
	return &agent{
		pilotAddress: address,
		node: &v3.Node{
			Id: opts.NodeID(),
			Metadata: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"ISTIO_VERSION": {
						Kind: &structpb.Value_StringValue{StringValue: "1.13.4"},
					},
				},
			},
		},
		opts:     opts,
		respChan: make(chan *discovery.DiscoveryResponse),
		log:      log.WithValues("nodeID", opts.NodeID()),
	}
}

// Send the request.
func (a *agent) Send(req *discovery.DiscoveryRequest) error {
	req.Node = a.node
	return a.stream.Send(req)
}

func (a *agent) WaitForRequestLimit(ctx context.Context) error {
	if a.connRateLimit == nil || a.connRateLimit.Limit() == 0 {
		// Allow opt out when rate limiting is set to 0qps
		return nil
	}
	// Give a bit of time for queue to clear out, but if not fail fast. Client will connect to another
	// instance in best case, or retry with backoff.
	wait, cancel := context.WithTimeout(ctx, time.Second*100)
	defer cancel()
	return a.connRateLimit.Wait(wait)
}

func (a *agent) Connect(ctx context.Context) error {
	err := a.WaitForRequestLimit(ctx)
	if err != nil {
		return err
	}
	conn, err := grpc.DialContext(ctx, a.pilotAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.Errorf("failed to connect to upstream %s: %v", a.pilotAddress, err)
		return err
	}
	a.conn = conn
	xds := discovery.NewAggregatedDiscoveryServiceClient(conn)
	a.stream, err = xds.StreamAggregatedResources(ctx,
		grpc.MaxCallRecvMsgSize(defaultClientMaxReceiveMessageSize))
	if err != nil {
		conn.Close()
		klog.Errorf("failed to create upstream grpc client: %v", err)
		return err
	}

	a.log.Info("connected to upstream XDS server.", "pilotAddress", a.pilotAddress)
	return nil
}

func (a *agent) Run(ctx context.Context, handler func(response *discovery.DiscoveryResponse)) {
	go func() {
		for {
			resp, err := a.stream.Recv()
			if err != nil {
				a.log.Error(err, "receive msg failed, exit...")
				break
			}
			a.respChan <- resp
		}
	}()
	go func() {
		select {
		case <-ctx.Done():
			a.log.Info("receive stop signal, terminal...")
			a.respChan <- nil
			return
		}
	}()
	for res := range a.respChan {
		if res == nil {
			break
		}
		handler(res)
	}
}
