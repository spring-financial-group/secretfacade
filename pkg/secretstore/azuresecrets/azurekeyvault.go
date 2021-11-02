package azuresecrets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	kvops "github.com/Azure/azure-sdk-for-go/services/keyvault/v7.1/keyvault"
	"github.com/jenkins-x-plugins/secretfacade/pkg/iam/azureiam"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore"
	"github.com/pkg/errors"
)

func NewAzureKeyVaultSecretManager(creds azureiam.Credentials) secretstore.Interface {
	return &azureKeyVaultSecretManager{creds}
}

type azureKeyVaultSecretManager struct {
	Creds azureiam.Credentials
}

func (a *azureKeyVaultSecretManager) GetSecret(vaultName, secretName, secretKey string) (string, error) {
	vaultURL, err := url.Parse(fmt.Sprintf("https://%s.vault.azure.net/", vaultName))
	if err != nil {
		return "", errors.Wrapf(err, "error getting secret for Azure Key Vault for secret %s from vault %s", secretName, vaultName)
	}
	keyClient, err := getSecretOpsClient(a.Creds)
	if err != nil {
		return "", errors.Wrap(err, "unable to create key ops client")
	}
	bundle, err := keyClient.GetSecret(context.TODO(), vaultURL.String(), secretName, "")
	if err != nil {
		return "", errors.Wrapf(err, "unable to retrieve secret %s from vault %s", secretName, vaultURL)
	}
	if bundle.Value == nil {
		return "", errors.Wrapf(err, "secret is empty for secret %s in vault %s", secretName, vaultURL)
	}
	var secretString string
	if secretKey != "" {
		secretString, err = getSecretProperty(bundle, secretKey)
		if err != nil {
			return "", errors.Wrapf(err, "error retrieving secret property from secret %s returned from Azure Key Vault %s", secretName, vaultName)
		}
	} else {
		secretString = *bundle.Value
	}
	return secretString, nil
}

func (a *azureKeyVaultSecretManager) SetSecret(vaultName, secretName string, secretValue *secretstore.SecretValue) error {
	vaultURL, err := url.Parse(fmt.Sprintf("https://%s.vault.azure.net/", vaultName))
	if err != nil {
		return errors.Wrapf(err, "error setting Azure Key Vault secret %s in vault %s", secretName, vaultName)
	}
	keyClient, err := getSecretOpsClient(a.Creds)
	if err != nil {
		return errors.Wrap(err, "unable to create key ops client")
	}
	secretString := secretValue.ToString()
	params := kvops.SecretSetParameters{
		Value: &secretString,
	}
	_, err = keyClient.SetSecret(context.TODO(), vaultURL.String(), secretName, params)

	if err != nil {
		return errors.Wrap(err, "unable to create key ops client")
	}

	return nil
}

func getSecretPropertyMap(v kvops.SecretBundle) (map[string]string, error) {
	m := make(map[string]string)
	secretString := *v.Value
	secretBytes := []byte(secretString)
	err := json.Unmarshal(secretBytes, &m)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling GCP secrets manager secret payload in to map[string]string")
	}
	return m, nil
}

func getSecretProperty(v kvops.SecretBundle, propertyName string) (string, error) {
	m, err := getSecretPropertyMap(v)
	if err != nil {
		return "", errors.Wrapf(err, "error reading property %s from secret JSON object", propertyName)
	}
	return m[propertyName], nil
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
