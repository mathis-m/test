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
	joinCluster := viper.GetBool("k8s-join-cluster")

	if initCluster {
		if err := runInitCluster(spank); err != nil {
			return err
		}
	} else if joinCluster {
		if err := runJoinCluster(spank); err != nil {
			return err
		}
	}

	return nil
}

func runJoinCluster(spank unsafe.Pointer) error {
	joinToken := viper.GetString("k8s-join-token")
	certHash := viper.GetString("k8s-join-cert-hash")
	apiEndpoint := viper.GetString("k8s-join-api-server")
	labels := viper.GetString("k8s-kublet-node-labels")
	taints := viper.GetString("k8s-kublet-node-taints")

	if err := setEnvironmentVariables(environmentVariables{
		Token:       joinToken,
		CaCertHash:  certHash,
		ApiEndpoint: apiEndpoint,
		Labels:      labels,
		Taints:      taints,
	}); err != nil {
		return err
	}

	return nil
}

func runInitCluster(spank unsafe.Pointer) error {
	joinToken, err := generateToken()
	if err != nil {
		return err
	}

	cert, err := generateCaCert()
	if err != nil {
		return err
	}

	if err := setEnvironmentVariables(environmentVariables{
		Token:      joinToken,
		CaCertHash: cert.CaCertHash,
		CaKeyB64:   cert.CaKeyB64,
		CaCertB64:  cert.CaCertB64,
	}); err != nil {
		return err
	}

	return nil
}

type environmentVariables struct {
	Token       string
	CaCertB64   string
	CaKeyB64    string
	CaCertHash  string
	ApiEndpoint string
	Labels      string
	Taints      string
}

func setEnvironmentVariables(vars environmentVariables) error {
	envVars := map[string]string{
		"SLURM_K8S_BOOTSTRAP_TOKEN": vars.Token,
		"SLURM_K8S_CA_CERT_HASH":    vars.CaCertHash,
		"SLURM_K8S_CA_CERT":         vars.CaCertB64,
		"SLURM_K8S_CA_KEY":          vars.CaKeyB64,
		"SLURM_K8S_API_ENDPOINT":    vars.ApiEndpoint,
		"SLURM_K8S_LABELS":          vars.Labels,
		"SLURM_K8S_TAINTS":          vars.Taints,
	}

	for key, value := range envVars {
		if value == "" {
			continue
		}

		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("unable to set environment variable %q: %w", key, err)
		}

		logValue := value
		if len(logValue) > 100 {
			logValue = logValue[:100] + "..."
		}

		log.Infof("%v: %v", key, logValue)
	}

	return nil
}

func generateToken() (string, error) {
	bootstrapToken, err := bootstraputil.GenerateBootstrapToken()
	if err != nil {
		return "", fmt.Errorf("unable to generate bootstrap token: %w", err)
	}

	return bootstrapToken, nil
}

type kubeCert struct {
	CaCertB64  string
	CaKeyB64   string
	CaCertHash string
}

func generateCaCert() (kubeCert, error) {
	homeDir := os.Getenv("HOME")

	certDir, err := os.MkdirTemp(homeDir, "tmp-ca")
	if err != nil {
		return kubeCert{}, err
	}

	cmdResult, err := util.RunCommand("kubeadm", "init", "phase", "certs", "ca", "--cert-dir", path.Join(homeDir, certDir))
	if err != nil {
		return kubeCert{}, err
	}
	if cmdResult.ExitCode != 0 {
		return kubeCert{}, fmt.Errorf("kubeadm failed with exit code: %v", cmdResult.ExitCode)
	}

	certPath := path.Join(homeDir, certDir, "ca.crt")
	keyPath := path.Join(homeDir, certDir, "ca.key")

	certContent, err := ioutil.ReadFile(certPath)
	if err != nil {
		return kubeCert{}, err
	}

	keyContent, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return kubeCert{}, err
	}

	if err := os.RemoveAll(path.Join(homeDir, certDir)); err != nil {
		return kubeCert{}, err
	}

	log.Infof(
		"tmp path: %v",
		certPath,
	)

	certB64 := base64.StdEncoding.EncodeToString(certContent)
	keyB64 := base64.StdEncoding.EncodeToString(keyContent)

	certHash, err := util.CalculatePublicKeyHash(certContent)
	if err != nil {
		return kubeCert{}, err
	}

	return kubeCert{
		CaCertB64:  certB64,
		CaKeyB64:   keyB64,
		CaCertHash: certHash,
	}, nil
}
