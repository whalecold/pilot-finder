package mock

import (
	"context"
	"sync"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/go-logr/logr"
	"golang.org/x/time/rate"
	"istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/whalecold/pilot-finder/pkg/agent"
	"github.com/whalecold/pilot-finder/pkg/constants"
	"github.com/whalecold/pilot-finder/pkg/kube"
	"github.com/whalecold/pilot-finder/pkg/utils/ip"
)

type workloadEntry struct {
	agent agent.Agent
	*networkingv1alpha3.WorkloadEntry
}

func initWorkloadEntry(name string, t *ip.Tools, i, j int, log logr.Logger, opts *Options) *workloadEntry {
	we := &workloadEntry{
		WorkloadEntry: &networkingv1alpha3.WorkloadEntry{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: opts.Namespace,
				Name:      name,
				Labels: map[string]string{
					"createdBy":    "pilot-finder",
					"serviceEntry": name,
				},
			},
			Spec: v1alpha3.WorkloadEntry{
				Address: t.IPv4(i, j).String(),
				Labels: map[string]string{
					"serviceEntry": name,
					"app":          name,
				},
				Ports: map[string]uint32{
					"http": 80,
				},
			},
		},
	}
	we.agent = agent.New(opts.PilotAddress, &agent.Options{
		Type:      "sidecar",
		Address:   t.IPv4(i, j).String(),
		PodName:   name,
		Namespace: opts.Namespace,
	}, log)
	return we
}

func (we *workloadEntry) init(ctx context.Context, cli client.Client, rateLimit *rate.Limiter) error {
	err := kube.CreateOrUpdate(ctx, cli, we.WorkloadEntry)
	if err != nil {
		return err
	}
	we.agent.SetConnectLimit(rateLimit)
	err = we.agent.Connect(ctx)
	if err != nil {
		return err
	}
	return we.agent.Send(&discovery.DiscoveryRequest{
		TypeUrl: "type.googleapis.com/envoy.config.listener.v3.Listener",
	})
}

func (we *workloadEntry) handler(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	we.agent.Run(ctx, func(response *discovery.DiscoveryResponse) {
		klog.V(constants.LevelGeneral).Infof("receive response url %s resource number %d", response.TypeUrl, len(response.Resources))
	})
}
