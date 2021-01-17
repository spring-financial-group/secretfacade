package vaultiam

import (
	"fmt"
	"os"
)

type VaultCreds struct {
	Token      string
	CaCertPath string
}

func NewEnvironmentCreds() (VaultCreds, error) {
	token, ok := os.LookupEnv("VAULT_TOKEN")
	if !ok || token == "" {
		return VaultCreds{}, fmt.Errorf("")
	}

	caCertPath := os.Getenv("VAULT_CACERT")

	return VaultCreds{
		Token:      token,
		CaCertPath: caCertPath,
	}, nil
}
