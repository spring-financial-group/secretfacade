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
	isExternalVault := os.Getenv("EXTERNAL_VAULT")

	// Skip the check for vault related environment variables if using external vault
	if isExternalVault == "true" {
		return VaultCreds{
			Token:      "",
			CaCertPath: "",
		}, nil
	}

	token, ok := os.LookupEnv("VAULT_TOKEN")
	if !ok || token == "" {
		return VaultCreds{}, fmt.Errorf("unable to find VAULT_TOKEN on environment")
	}

	caCertPath := os.Getenv("VAULT_CACERT")

	return VaultCreds{
		Token:      token,
		CaCertPath: caCertPath,
	}, nil
}
