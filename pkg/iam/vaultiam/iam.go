package vaultiam

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	secretNamespace       = "secret-infra"
	externalSecretsPrefix = "kubernetes-external-secrets-token"
)

type VaultCreds struct {
	Token      string
	CaCertPath string
}

func NewEnvironmentCreds() (VaultCreds, error) {

	token, ok := os.LookupEnv("VAULT_TOKEN")
	// Skip the check for vault related environment variables if using external vault
	if !ok || token == "" {
		return VaultCreds{}, fmt.Errorf("unable to find VAULT_TOKEN on environment")
	}

	caCertPath := os.Getenv("VAULT_CACERT")

	return VaultCreds{
		Token:      token,
		CaCertPath: caCertPath,
	}, nil
}

func NewExternalSecretCreds(client *api.Client, kubeClient kubernetes.Interface) (VaultCreds, error) {
	token, err := getTokenForExternalVault(client, kubeClient)
	if err != nil {
		return VaultCreds{}, errors.Wrap(err, "error getting client token for external vault")
	}

	caCertPath := os.Getenv("VAULT_CACERT")

	return VaultCreds{
		Token:      token,
		CaCertPath: caCertPath,
	}, nil
}

// Taken from https://www.vaultproject.io/docs/auth/kubernetes#code-example
func getTokenForExternalVault(client *api.Client, kClient kubernetes.Interface) (string, error) {
	vaultMountPoint := os.Getenv("JX_VAULT_MOUNT_POINT")
	if vaultMountPoint == "" {
		vaultMountPoint = "kubernetes"
		log.Logger().Infof("Setting vault mount point to kubernetes as JX_VAULT_MOUNT_POINT is missing")
	}
	vaultRole := os.Getenv("JX_VAULT_ROLE")
	if vaultRole == "" {
		vaultRole = "jx-vault"
		log.Logger().Infof("Setting vault role to jx-vault as JX_VAULT_ROLE is missing")
	}

	secrets, err := kClient.CoreV1().Secrets(secretNamespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", errors.Wrap(err, "error listing secrets")
	}
	var secretName string
	for _, v := range secrets.Items {
		if strings.HasPrefix(v.Name, externalSecretsPrefix) {
			secretName = v.Name
			break
		}
	}

	secret, err := kClient.CoreV1().Secrets(secretNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return "", errors.Wrap(err, "error listing secrets")
	}

	token := string(secret.Data["token"])

	params := map[string]interface{}{
		"jwt":  token,
		"role": vaultRole,
	}
	// log in to Vault's Kubernetes auth method
	resp, err := client.Logical().Write("auth/"+vaultMountPoint+"/login", params)
	if err != nil {
		return "", errors.Wrap(err, "unable to log in with Kubernetes auth")
	}
	if resp == nil || resp.Auth == nil || resp.Auth.ClientToken == "" {
		return "", fmt.Errorf("login response did not return client token")
	}
	return resp.Auth.ClientToken, nil
}
