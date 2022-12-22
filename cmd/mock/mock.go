package mock

import (
	"context"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"flag"

	"github.com/whalecold/pilot-finder/pkg/mock"
)

func NewCommand(ctx context.Context) *cobra.Command {
	klog.InitFlags(nil)
	flag.Parse()
	opts := &mock.Options{}
	cmd := &cobra.Command{
		Use:          "mock",
		Short:        "mock the multi agents connecting to pilot",
		SilenceUsage: true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			// Allow unknown flags for backward-compatibility.
			UnknownFlags: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			klog.Infoln("opts ", opts)
			m := mock.New(opts, klog.NewKlogr())
			m.Init()
			return m.Run(ctx)
		},
	}
	cmd.Flags().IntVarP(&opts.ServiceEntryNumber, "serviceEntryNumber", "s", 1, "serviceEntry number")
	cmd.Flags().IntVarP(&opts.WorkloadEntryNumber, "workloadEntryNumber", "w", 1, "workloadEntry number per serviceEntry")
	cmd.Flags().StringVarP(&opts.AddressPrefix, "addressPrefix", "a", "192.168", "the address prefix, use for generator the address of workload entry")
	cmd.Flags().StringVarP(&opts.PilotAddress, "pilotAddress", "p", "127.0.0.1:15010", "specify the pilot-discovery address")
	cmd.Flags().StringVarP(&opts.Namespace, "namespace", "n", "default", "specify the namespace where se located.")
	cmd.Flags().Float64Var(&opts.ConnectRateLimit, "rate", 10, "specify rate limit for connect to pilot.")
	cmd.Flags().Float32Var(&opts.Options.Qps, "qps", 25, "the maximum QPS of the Kubernetes client.")
	cmd.Flags().IntVar(&opts.Options.Burst, "burst", 50, "maximum burst for throttle.")
	return cmd
}
