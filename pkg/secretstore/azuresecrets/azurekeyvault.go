package azuresecrets

import (
	"context"
	"fmt"
	"net/url"

	kvops "github.com/Azure/azure-sdk-for-go/services/keyvault/v7.1/keyvault"
	"github.com/chrismellard/secretfacade/pkg/iam/azureiam"
	"github.com/chrismellard/secretfacade/pkg/secretstore"
	"github.com/pkg/errors"
)

func NewAzureKeyVaultSecretManager(creds azureiam.Credentials) secretstore.Interface {
	return &azureKeyVaultSecretManager{creds}
}

type azureKeyVaultSecretManager struct {
	Creds azureiam.Credentials
}

func (a *azureKeyVaultSecretManager) GetSecret(vaultName string, secretName string, _ string) (string, error) {
	vaultUrl, err := url.Parse(fmt.Sprintf("https://%s.vault.azure.net/", vaultName))
	if err != nil {
		return "", errors.Wrap(err, "")
	}
	keyClient, err := getSecretOpsClient(a.Creds)
	if err != nil {
		return "", errors.Wrap(err, "unable to create key ops client")
	}
	bundle, err := keyClient.GetSecret(context.TODO(), vaultUrl.String(), secretName, "")
	if err != nil {
		return "", errors.Wrapf(err, "unable to retrieve secret %s from vault %s", secretName, vaultUrl)
	}
	if bundle.Value == nil {
		return "", errors.Wrapf(err, "secret is empty for secret %s in vault %s", secretName, vaultUrl)
	}
	return *bundle.Value, nil
}

func (a *azureKeyVaultSecretManager) SetSecret(vaultName string, secretName string, secretValue *secretstore.SecretValue) error {
	vaultUrl, err := url.Parse(fmt.Sprintf("https://%s.vault.azure.net/", vaultName))
	if err != nil {
		return errors.Wrap(err, "")
	}
	keyClient, err := getSecretOpsClient(a.Creds)
	if err != nil {
		return errors.Wrap(err, "unable to create key ops client")
	}
	secretString := secretValue.ToString()
	params := kvops.SecretSetParameters{
		Value: &secretString,
	}
	_, err = keyClient.SetSecret(context.TODO(), vaultUrl.String(), secretName, params)

	if err != nil {
		return errors.Wrap(err, "unable to create key ops client")
	}

	return nil
}

func getSecretOpsClient(creds azureiam.Credentials) (*kvops.BaseClient, error) {
	keyvaultClient := kvops.New()
	a, err := azureiam.GetKeyvaultAuthorizer(creds)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create key vault authorizer")
	}
	keyvaultClient.Authorizer = a
	return &keyvaultClient, nil
}
