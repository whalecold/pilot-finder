package xds

import (
	"context"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/spf13/cobra"
	"github.com/whalecold/pilot-finder/pkg/agent"
	"k8s.io/klog/v2"
)

type options struct {
	url                   string
	pilotDiscoveryAddress string
	agent.Options
}

var (
	xdsTypeMap = map[string]string{
		"lds": "type.googleapis.com/envoy.config.listener.v3.Listener",
	}
)

func NewCommand(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:          "xds",
		Short:        "take the request from the pilot discovery",
		SilenceUsage: true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			// Allow unknown flags for backward-compatibility.
			UnknownFlags: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			klog.Infoln("opts ", opts)
			cli := agent.New(opts.pilotDiscoveryAddress, &opts.Options, klog.NewKlogr())
			err := cli.Connect(context.Background())
			if err != nil {
				return err
			}
			defer cli.Close()

			cli.Send(&discovery.DiscoveryRequest{
				TypeUrl: xdsTypeMap[opts.url],
			})
			cli.Run(ctx, func(response *discovery.DiscoveryResponse) {
				klog.Infof("receive the resp url %s", response.TypeUrl)
				return
			})
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.url, "url", "lds", "specify the url type of xds, options for lds|rds|cds")
	cmd.Flags().StringVar(&opts.pilotDiscoveryAddress, "discoveryAddress", "127.0.0.1:15010", "specify the pilot-discovery address")
	cmd.Flags().StringVar(&opts.Type, "agent", "sidecar", "agent type")
	cmd.Flags().StringVar(&opts.Address, "ip", "127.0.0.1", "the address of the agent")
	cmd.Flags().StringVar(&opts.PodName, "podName", "pod", "the agent name")
	cmd.Flags().StringVar(&opts.Namespace, "namespace", "default", "the agent location namespace")
	cmd.Flags().StringVar(&opts.Suffix, "suffix", "svc.cluster.local", "the cluster domain suffix")
	return cmd
}