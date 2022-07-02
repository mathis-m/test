package cmd

import (
	"github.com/s-bauer/slurm-k8s/internal/useful_paths"
	"github.com/s-bauer/slurm-k8s/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		service := &util.Service{Name: useful_paths.ServicesRootlesskit}

		if err := util.ReloadSystemdDaemon(); err != nil {
			log.Fatalf("unable to reload systemd daemon: %v", err)
		}

		status, err := service.Status()
		if err != nil {
			log.Fatal(err)
		}

		if status == util.Active {
			if viper.GetBool("restart") {
				log.Infof("service %q is running, stopping first...", service.Name)
				if err := service.Stop(); err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatalf("service %q is running, stop first or include --restart flag", service.Name)
			}
		}

		if err := service.Start(); err != nil {
			log.Fatal(err)
		}

		log.Infof("started service %q", service.Name)

		usefulPaths, err := useful_paths.ConstructUsefulPaths()
		if err != nil {
			log.Fatal(err)
		}

		log.Info("starting bootstrap process - this will be blocking")
		cmdResult, err := util.RunCommand(usefulPaths.Scripts.BootstrapCluster)
		if err != nil || cmdResult.ExitCode != 0 {
			log.Fatalf("failed to bootstrap-cluster: %v", err)
		}

	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().BoolP("restart", "r", false, "restarts the service if it's already running")

	_ = viper.BindPFlags(startCmd.Flags())
}
