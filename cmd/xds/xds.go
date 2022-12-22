package xds

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/whalecold/pilot-finder/cmd/xds/client"
	"github.com/whalecold/pilot-finder/cmd/xds/server"
)

func NewCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use: "xds",
	}
	cmd.AddCommand(client.NewCommand(ctx))
	cmd.AddCommand(server.NewCommand(ctx))
	return cmd
}
