package server

import (
	"context"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/whalecold/pilot-finder/pkg/server"
)

type options struct {
	port int
}

var (
	xdsTypeMap = map[string]string{
		"lds": "type.googleapis.com/envoy.config.listener.v3.Listener",
	}
)

func NewCommand(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:          "server",
		Short:        "mock the xds server to test",
		SilenceUsage: true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			// Allow unknown flags for backward-compatibility.
			UnknownFlags: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			klog.Infoln("opts ", opts)
			_, err := server.New(opts.port, klog.NewKlogr())
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&opts.port, "port", 9099, "specify the port")
	return cmd
}
