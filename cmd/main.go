package main

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apiserver/pkg/server"

	"github.com/whalecold/pilot-finder/cmd/mock"
	"github.com/whalecold/pilot-finder/cmd/xds"
	"github.com/whalecold/pilot-finder/pkg/version"
	"k8s.io/klog/v2"
)

var (
	rootCmd = &cobra.Command{
		Use:          "pilot-finder",
		Short:        "pilot finder is a test tool for pilot-discovery",
		Version:      version.Version,
		SilenceUsage: true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			// Allow unknown flags for backward-compatibility.
			UnknownFlags: true,
		},
	}
)

func init() {
	ctx := server.SetupSignalContext()
	rootCmd.AddCommand(xds.NewCommand(ctx))
	rootCmd.AddCommand(mock.NewCommand(ctx))
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	res := rootCmd.Execute()
	if res != nil {
		os.Exit(1)
	}
}
