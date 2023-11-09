package vm

import (
	"github.com/mikrolite/mikrolite/adapters/vm/firecracker"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func NewVMCommand() *cobra.Command {
	cfg := &commonConfig{}

	cmd := &cobra.Command{
		Use:   "vm",
		Short: "Create and manage virtual machines",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			loggerOpts := &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}
			if cfg.Debug {
				loggerOpts.Level = slog.LevelDebug
			}
			logger := slog.New(slog.NewTextHandler(os.Stdout, loggerOpts))
			slog.SetDefault(logger)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringVar(&cfg.SocketPath, "socket-path", "/run/containerd/containerd.sock", "the path to the containerd socket")
	cmd.PersistentFlags().StringVar(&cfg.StateRootPath, "state-path", "/usr/local/share/mikrolite", "the path to the root directory to hold state in")
	cmd.PersistentFlags().BoolVar(&cfg.Debug, "debug", false, "enable debug features")
	cmd.PersistentFlags().StringVarP(&cfg.VMProvider, "provider", "p", firecracker.ProviderName, "the vm provider to use")
	cmd.PersistentFlags().StringVar(&cfg.FirecrackerBin, "firecracker-bin", "firecracker", "the path to the firecracker binary to use")
	cmd.PersistentFlags().StringVar(&cfg.CloudHypervisorBin, "cloudhypervisor-bin", "cloud-hypervisor-static", "the path to the cloud-hypervisor binary to use")

	cmd.AddCommand(newCreateCommandVM(cfg))
	cmd.AddCommand(newRemoveVMCommand(cfg))

	return cmd
}

type commonConfig struct {
	SocketPath         string
	StateRootPath      string
	Debug              bool
	VMProvider         string
	FirecrackerBin     string
	CloudHypervisorBin string
}
