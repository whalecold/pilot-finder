package mock

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/whalecold/pilot-finder/pkg/constants"
	"github.com/whalecold/pilot-finder/pkg/kube"
	"github.com/whalecold/pilot-finder/pkg/utils/ip"
	"golang.org/x/time/rate"
)

type serviceEntry struct {
	wes []*workloadEntry
	*networkingv1alpha3.ServiceEntry
}

func initServiceEntry(idx int, t *ip.Tools, log logr.Logger, opts *Options) *serviceEntry {
	name := fmt.Sprintf("serviceentry-%03d", idx)
	se := &serviceEntry{
		ServiceEntry: &networkingv1alpha3.ServiceEntry{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: opts.Namespace,
				Name:      name,
				Labels: map[string]string{
					"createdBy": "pilot-finder",
				},
			},
			Spec: v1alpha3.ServiceEntry{
				Hosts: []string{
					name + "." + opts.Namespace + "." + constants.NodeIDSuffix,
				},
				Ports: []*v1alpha3.Port{
					{
						Name:     "http",
						Number:   80,
						Protocol: "HTTP",
					},
				},
				Resolution: v1alpha3.ServiceEntry_STATIC,
				WorkloadSelector: &v1alpha3.WorkloadSelector{
					Labels: map[string]string{
						"app": name,
					},
				},
			},
		},
	}

	for j := 0; j < opts.WorkloadEntryNumber; j++ {
		wname := fmt.Sprintf("%s-workload-%03d", name, j)
		we := initWorkloadEntry(wname, t, idx, j, log, opts)
		se.wes = append(se.wes, we)
	}
	return se
}

func (se *serviceEntry) run(ctx context.Context, cli client.Client, wg *sync.WaitGroup, rateLimit *rate.Limiter) error {
	err := kube.CreateOrUpdate(ctx, cli, se.ServiceEntry)
	if err != nil {
		return err
	}
	for _, we := range se.wes {
		err = we.init(ctx, cli, rateLimit)
		if err != nil {
			return err
		}
		wg.Add(1)
		go we.handler(ctx, wg)
	}
	return nil
}
