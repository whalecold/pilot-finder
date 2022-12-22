package mock

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/whalecold/pilot-finder/pkg/kube"
	"github.com/whalecold/pilot-finder/pkg/utils/ip"
	"golang.org/x/time/rate"
)

// Options options to run
type Options struct {
	// ServiceEntryNumber service entry number
	ServiceEntryNumber int
	// WorkloadEntryNumber the workload entry number per serviceEntry.
	WorkloadEntryNumber int
	// AddressPrefix address predix, eg "10.180"
	AddressPrefix string
	// PilotAddress the pilot address.
	PilotAddress string
	// Namespace ...
	Namespace string
	// ConnectRateLimit limits the number of new XDS requests allowed. This helps prevent thundering hurd of connect
	// requests at same time.
	ConnectRateLimit float64

	kube.Options
}

type Interface interface {
	Run(ctx context.Context) error
	Init() error
}

type mock struct {
	cli  client.Client
	ses  []*serviceEntry
	opts *Options

	// limits the number of new XDS requests allowed. This helps prevent thundering hurd of connect
	// requests at same time.
	connRateLimit *rate.Limiter
	log           logr.Logger
}

func (m *mock) Init() error {
	cli, err := kube.NewClient(&m.opts.Options)
	if err != nil {
		return err
	}
	m.cli = cli

	klog.Info("init kube client success")

	m.connRateLimit = rate.NewLimiter(rate.Limit(m.opts.ConnectRateLimit), 1)

	t := ip.New(m.opts.AddressPrefix)
	if t == nil {
		return fmt.Errorf("init address failed %s", m.opts.AddressPrefix)
	}
	m.ses = make([]*serviceEntry, 0, m.opts.ServiceEntryNumber)
	for i := 0; i < m.opts.ServiceEntryNumber; i++ {
		se := initServiceEntry(i, t, m.log, m.opts)
		m.ses = append(m.ses, se)
	}
	return nil
}

func (m *mock) Run(ctx context.Context) error {
	wg := sync.WaitGroup{}
	for _, se := range m.ses {
		err := se.run(ctx, m.cli, &wg, m.connRateLimit)
		if err != nil {
			return err
		}
	}
	wg.Wait()
	return nil
}

func New(opts *Options, log logr.Logger) Interface {
	return &mock{
		log:  log,
		opts: opts,
	}
}
