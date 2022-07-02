package main

import (
	"github.com/s-bauer/slurm-k8s/cmd/bootstrap-uk8s/cmd"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
}

func main() {
	log.Info("bootstrap-uk8s starting")
	cmd.Execute()
}
