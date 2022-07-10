package spank_local

import "C"
import (
	"encoding/base64"
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
	"os"
	"path"
	"unsafe"
)

func Init(spank unsafe.Pointer) error {
	initCluster := viper.GetBool("k8s-init-cluster")

	if initCluster {
		if err := runInitCluster(); err != nil {
			return err
		}
	}

	return nil
}

func runInitCluster() error {
	if err := generateToken(); err != nil {
		return err
	}

	if err := generateCaCert(); err != nil {
		return err
	}

	return nil
}

func generateToken() error {
	bootstrapToken, err := bootstraputil.GenerateBootstrapToken()
	if err != nil {
		return fmt.Errorf("unable to generate bootstrap token: %w", err)
	}

	if err := os.Setenv("SLURM_K8S_BOOTSTRAP_TOKEN", bootstrapToken); err != nil {
		return fmt.Errorf("unable to set environment variable: %w", err)
	}

	log.Infof("SLURM_K8S_BOOTSTRAP_TOKEN: %v", bootstrapToken)

	return nil
}

func generateCaCert() error {
	homeDir := os.Getenv("HOME")

	certDir, err := os.MkdirTemp(homeDir, "tmp-ca")
	if err != nil {
		return err
	}

	cmdResult, err := util.RunCommand("kubeadm", "init", "phase", "certs", "ca", "--cert-dir", path.Join(homeDir, certDir))
	if err != nil {
		return err
	}
	if cmdResult.ExitCode != 0 {
		return fmt.Errorf("kubeadm failed with exit code: %v", cmdResult.ExitCode)
	}

	certPath := path.Join(homeDir, certDir, "ca.crt")
	keyPath := path.Join(homeDir, certDir, "ca.key")

	certContent, err := ioutil.ReadFile(certPath)
	if err != nil {
		return err
	}

	keyContent, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(path.Join(homeDir, certDir)); err != nil {
		return err
	}

	log.Infof(
		"tmp path: %v",
		certPath,
	)

	certB64 := base64.StdEncoding.EncodeToString(certContent)
	keyB64 := base64.StdEncoding.EncodeToString(keyContent)

	certHash, err := util.CalculatePublicKeyHash(certContent)
	if err != nil {
		return err
	}

	// log
	log.Infof("SLURM_K8S_CA_CERT: %v", certB64)
	log.Infof("SLURM_K8S_CA_KEY: %v", keyB64)
	log.Infof("SLURM_K8S_CA_CERT_HASH=%v", certHash)

	// set environment variables
	if err := os.Setenv("SLURM_K8S_CA_CERT", certB64); err != nil {
		return fmt.Errorf("unable to set environment variable: %w", err)
	}

	if err := os.Setenv("SLURM_K8S_CA_KEY", keyB64); err != nil {
		return fmt.Errorf("unable to set environment variable: %w", err)
	}

	if err := os.Setenv("SLURM_K8S_CA_CERT_HASH", certHash); err != nil {
		return fmt.Errorf("unable to set environment variable: %w", err)
	}

	return nil
}
