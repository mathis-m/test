package util

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
)

func CalculatePublicKeyHash(certificate []byte) (string, error) {
	certBlock, _ := pem.Decode(certificate)
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return "", fmt.Errorf("unable to parse x509 certificate: %w", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		return "", fmt.Errorf("unable to marshal public key: %w", err)
	}

	hasher := sha256.New()
	hasher.Write(publicKeyBytes)
	sha256Hex := hex.EncodeToString(hasher.Sum(nil))

	return sha256Hex, nil
}
