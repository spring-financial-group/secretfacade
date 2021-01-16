package azuresecrets

import (
	"context"

	kvops "github.com/Azure/azure-sdk-for-go/services/preview/keyvault/v7.2-preview/keyvault"
	"github.com/chrismellard/secretfacade/pkg/iam/azure"
	"github.com/chrismellard/secretfacade/pkg/secretstore"
	"github.com/pkg/errors"
)

type AzureKeyVaultSecretManager struct {
	Creds azure.Credentials
}

func (a *AzureKeyVaultSecretManager) GetSecret(vaultUrl string, secretName string, _ string) (string, error) {
	keyClient, err := getSecretOpsClient(a.Creds)
	if err != nil {
		return "", errors.Wrap(err, "unable to create key ops client")
	}
	bundle, err := keyClient.GetSecret(context.TODO(), vaultUrl, secretName, "")
	if err != nil {
		return "", errors.Wrapf(err, "unable to retrieve secret %s from vault %s", secretName, vaultUrl)
	}
	if bundle.Value == nil {
		return "", errors.Wrapf(err, "secret is empty for secret %s in vault %s", secretName, vaultUrl)
	}
	return *bundle.Value, nil
}

func (a *AzureKeyVaultSecretManager) SetSecret(vaultUrl string, secretName string, secretValue *secretstore.SecretValue) error {
	keyClient, err := getSecretOpsClient(a.Creds)
	if err != nil {
		return errors.Wrap(err, "unable to create key ops client")
	}
	secretString := secretValue.ToString()
	params := kvops.SecretSetParameters{
		Value: &secretString,
	}
	_, err = keyClient.SetSecret(context.TODO(), vaultUrl, secretName, params)

	if err != nil {
		return errors.Wrap(err, "unable to create key ops client")
	}

	return nil
}

func getSecretOpsClient(creds azure.Credentials) (*kvops.BaseClient, error) {
	keyvaultClient := kvops.New()
	a, err := azure.GetKeyvaultAuthorizer(creds)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create key vault authorizer")
	}
	keyvaultClient.Authorizer = a
	return &keyvaultClient, nil
}
