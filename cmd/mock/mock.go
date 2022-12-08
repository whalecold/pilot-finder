package mock

import (
	"context"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/whalecold/pilot-finder/pkg/mock"
)

func NewCommand(ctx context.Context) *cobra.Command {
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
	cmd.Flags().IntVar(&opts.ServiceEntryNumber, "sn", 1, "serviceEntry number")
	cmd.Flags().IntVar(&opts.WorkloadEntryNumber, "wn", 1, "workloadEntry number per serviceEntry")
	cmd.Flags().StringVar(&opts.AddressPrefix, "ap", "192.168", "the address prefix, use for generator the address of workload entry")
	cmd.Flags().StringVar(&opts.PilotAddress, "pilotAddress", "127.0.0.1:15010", "specify the pilot-discovery address")
	cmd.Flags().StringVar(&opts.Namespace, "namespace", "default", "specify the namespace where se located.")
	cmd.Flags().Float64Var(&opts.ConnectRateLimit, "rate", 10, "specify rate limit for connect to pilot.")
	return cmd
}
