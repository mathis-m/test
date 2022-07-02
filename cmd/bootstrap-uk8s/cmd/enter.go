package cmd

import (
	"context"
	"github.com/s-bauer/slurm-k8s/internal/useful_paths"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/exec"

	"github.com/rootless-containers/rootlesskit/pkg/sigproxy"
	"github.com/rootless-containers/rootlesskit/pkg/sigproxy/signal"
)

// enterCmd represents the start command
var enterCmd = &cobra.Command{
	Use:   "enter",
	Short: "Enters the kubernetes namespace ",
	Run: func(cmd *cobra.Command, args []string) {
		usefulPaths, err := useful_paths.ConstructUsefulPaths()
		if err != nil {
			log.Fatal(err)
		}

		command := exec.Command(usefulPaths.Scripts.Nsenter)
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Start(); err != nil {
			log.Fatalf("failed to start the child: %v", err)
		}

		sigc := sigproxy.ForwardAllSignals(context.TODO(), command.Process.Pid)
		defer signal.StopCatch(sigc)

		// block until the child exits
		if err := command.Wait(); err != nil {
			log.Fatalf("child exited: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(enterCmd)
}
