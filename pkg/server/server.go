package server

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/go-logr/logr"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	mcp "istio.io/api/mcp/v1alpha1"
	networking "istio.io/api/networking/v1alpha3"
	"istio.io/pkg/env"
	"istio.io/pkg/log"
	istioversion "istio.io/pkg/version"
)

// IstioControlPlaneInstance defines the format Istio uses for when creating Envoy config.core.v3.ControlPlane.identifier
type IstioControlPlaneInstance struct {
	// The Istio component type (e.g. "istiod")
	Component string
	// The ID of the component instance
	ID string
	// The Istio version
	Info istioversion.BuildInfo
}

var controlPlane *corev3.ControlPlane

// ControlPlane identifies the instance and Istio version.
func ControlPlane() *corev3.ControlPlane {
	return controlPlane
}

func init() {
	podName := env.RegisterStringVar("POD_NAME", "", "").Get()
	byVersion, err := json.Marshal(IstioControlPlaneInstance{
		Component: "istiod",
		ID:        podName,
		Info:      istioversion.Info,
	})
	if err != nil {
		fmt.Println("XDS: Could not serialize control plane id:", err)
	}
	controlPlane = &corev3.ControlPlane{Identifier: string(byVersion)}
}

type Server struct {
	insecureGrpcServer *grpc.Server
	log                logr.Logger
}

func New(port int, log logr.Logger) (*Server, error) {

	s := &Server{
		log: log,
	}
	s.insecureGrpcServer = grpc.NewServer()
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	ke := os.Getenv("Ke")
	fmt.Println(ke)
	log.Info("start listen the server...", "port", port)
	discovery.RegisterAggregatedDiscoveryServiceServer(s.insecureGrpcServer, s)

	err = s.insecureGrpcServer.Serve(grpcListener)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Server) DeltaAggregatedResources(stream discovery.AggregatedDiscoveryService_DeltaAggregatedResourcesServer) error {
	return s.StreamDeltas(stream)
}
func (s *Server) StreamAggregatedResources(stream discovery.AggregatedDiscoveryService_StreamAggregatedResourcesServer) error {
	return nil
}

func nonce() string {
	return uuid.New().String()
}

type DeltaStream discovery.AggregatedDiscoveryService_DeltaAggregatedResourcesServer

func (s *Server) send(stream DeltaStream, typeURL string, res []*discovery.Resource, log logr.Logger) {
	resp := &discovery.DeltaDiscoveryResponse{
		ControlPlane: ControlPlane(),
		TypeUrl:      typeURL,
		Nonce:        nonce(),
		Resources:    res,
	}

	err := stream.Send(resp)
	if err != nil {
		log.Error(err, "send resp failed")
	}
}

type SendQueue struct {
	res []*discovery.Resource
	url string
}

func (s *Server) StreamDeltas(stream DeltaStream) error {
	ctx := stream.Context()
	peerAddr := "0.0.0.0"
	if peerInfo, ok := peer.FromContext(ctx); ok {
		peerAddr = peerInfo.Addr.String()
	}
	log := s.log.WithValues("peerAddr", peerAddr)
	log.Info("receive a new steam..")

	ch := make(chan *SendQueue, 10000)
	go s.Send(stream, ch, log)
	for {
		req, err := stream.Recv()
		if err != nil {
			log.Error(err, "receive from stream failed.")
			return err
		}
		if req.ResponseNonce != "" {
			log.Info("receive an ack msg", "type", req.TypeUrl, "nonce", req.ResponseNonce)
			continue
		}
		log.Info("receive info", "type", req.TypeUrl, "node", req.Node)
		switch req.TypeUrl {
		case "networking.istio.io/v1alpha3/ServiceEntry":
			// s.sendServiceEntry(ch, req.TypeUrl, log)
			//s.send(stream, req.TypeUrl, fakeServiceEntryResources(), log)
			ch <- &SendQueue{
				url: req.TypeUrl,
				res: fakeServiceEntryResources(),
			}
		case "networking.istio.io/v1alpha3/WorkloadEntry":
			//s.send(stream, req.TypeUrl, fakeWorkloadEntryResources(), log)
			ch <- &SendQueue{
				url: req.TypeUrl,
				res: fakeWorkloadEntryResources(),
			}
		}
	}
}

func (s *Server) Send(stream DeltaStream, ch chan *SendQueue, log logr.Logger) {
	for resp := range ch {
		s.send(stream, resp.url, resp.res, log)
		log.Info("send response", "type", resp.url)
	}
}

func inetntoa(ip int) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func (s *Server) sendServiceEntry(ch chan *SendQueue, typeURL string, log logr.Logger) {

	// for j := 0; j < 50000; j++ {
	//     for i := 0; i < 1; i++ {
	//         log.Info("send the number %d response..", "index", i)
	//         ch <- &SendQueue{
	//             url: typeURL,
	//             res: []*discovery.Resource{serviceEntryNumber(j*1000 + i)},
	//         }
	//     }
	// }
	log.Info("send number response 10000 completed..")
}

func workloadEntryNumber(i int) *discovery.Resource {
	suffix := fmt.Sprintf("%06d", i)
	se := &networking.WorkloadEntry{
		ServiceAccount: "servicea-000499",
		Labels: map[string]string{
			suffix: suffix,
		},
	}
	mcpRes := &mcp.Resource{
		Metadata: &mcp.Metadata{
			Name: fmt.Sprintf("test-auth-public-default-group/servicea-%s", suffix),
			CreateTime: &timestamp.Timestamp{
				Nanos: int32(time.Now().Nanosecond()),
			},
			Labels: map[string]string{
				"registrysynctask.mse.paas.volcengine.com/workspaceId": "qbmsqu0j4cd1zr5uoe7s9uqr",
			},
		},
		Body: MessageToAny(se),
	}

	return &discovery.Resource{
		Resource: MessageToAny(mcpRes),
	}
}

func serviceEntryNumber(i int) *discovery.Resource {
	suffix := fmt.Sprintf("%06d", i)
	se := &networking.ServiceEntry{
		Hosts: []string{
			fmt.Sprintf("servicea-%s.test-auth-public-default-group.svc.cluster.local", suffix),
		},
		Ports: []*networking.Port{
			{
				Name:     "http",
				Number:   80,
				Protocol: "HTTP",
			},
		},
		Resolution: networking.ServiceEntry_STATIC,
		Endpoints: []*networking.WorkloadEntry{
			{
				Address: inetntoa(i + 3221225473),
				Ports: map[string]uint32{
					"http": 80,
				},
			},
		},
	}
	mcpRes := &mcp.Resource{
		Metadata: &mcp.Metadata{
			Name: fmt.Sprintf("test-auth-public-default-group/servicea-%s", suffix),
			CreateTime: &timestamp.Timestamp{
				Nanos: int32(time.Now().Nanosecond()),
			},
			Labels: map[string]string{
				"registrysynctask.mse.paas.volcengine.com/workspaceId": "qbmsqu0j4cd1zr5uoe7s9uqr",
			},
		},
		Body: MessageToAny(se),
	}

	return &discovery.Resource{
		Resource: MessageToAny(mcpRes),
	}
}

func fakeWorkloadEntryResources() []*discovery.Resource {
	number := 5000
	ret := make([]*discovery.Resource, 0, number)
	for i := 0; i < number; i++ {
		ret = append(ret, workloadEntryNumber(i))
	}
	return ret
}

func fakeServiceEntryResources() []*discovery.Resource {
	number := 50000
	ret := make([]*discovery.Resource, 0, number)
	for i := 0; i < number; i++ {
		ret = append(ret, serviceEntryNumber(i))
	}
	return ret
}

// MessageToAnyWithError converts from proto message to proto Any
func MessageToAnyWithError(msg proto.Message) (*anypb.Any, error) {
	b, err := proto.MarshalOptions{Deterministic: true}.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return &anypb.Any{
		// nolint: staticcheck
		TypeUrl: "type.googleapis.com/" + string(msg.ProtoReflect().Descriptor().FullName()),
		Value:   b,
	}, nil
}

// MessageToAny converts from proto message to proto Any
func MessageToAny(msg proto.Message) *anypb.Any {
	out, err := MessageToAnyWithError(msg)
	if err != nil {
		log.Error(fmt.Sprintf("error marshaling Any %s: %v", prototext.Format(msg), err))
		return nil
	}
	return out
}
