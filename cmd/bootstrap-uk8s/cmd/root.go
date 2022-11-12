package cmd

import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"syscall"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bootstrap-uk8s",
	Short: "An utility to boostrap kubernetes in a linux namespace",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("verbose") {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}

		if viper.GetBool("simple-log") {
			log.SetFormatter(&log.TextFormatter{
				DisableColors:    true,
				DisableQuote:     true,
				DisableTimestamp: true,
			})
		}

		if !util.IsInNamespace() {
			intendedUid := viper.GetInt("drop-uid")
			intendedGid := viper.GetInt("drop-gid")

			oldUid := syscall.Getuid()
			oldGid := syscall.Getgid()

			log.Infof("previous uid=%v, gid=%v", oldUid, oldGid)

			if intendedUid >= 0 {
				if err := syscall.Setreuid(intendedUid, intendedUid); err != nil {
					log.Fatalf("unable to drop uid %v: %v", intendedUid, err)
				}
			}

			if intendedGid >= 0 {
				if err := syscall.Setregid(intendedGid, intendedGid); err != nil {
					log.Fatalf("unable to drop gid %v: %v", intendedGid, err)
				}
			}

			newUid := syscall.Getuid()
			newGid := syscall.Getgid()

			log.Infof("new uid=%v, gid=%v", newUid, newGid)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().Bool("simple-log", false, "disabled fancy log formatting")

	rootCmd.PersistentFlags().Int("drop-uid", -1, "drops the real and effective uid")
	rootCmd.PersistentFlags().Int("drop-gid", -1, "drops the real and effective gid")

	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
