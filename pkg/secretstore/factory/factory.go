package factory

import (
	"fmt"

	"github.com/chrismellard/secretfacade/pkg/iam/azureiam"
	gcpiam "github.com/chrismellard/secretfacade/pkg/iam/gcp"
	"github.com/chrismellard/secretfacade/pkg/secretstore"
	"github.com/chrismellard/secretfacade/pkg/secretstore/azuresecrets"
	"github.com/chrismellard/secretfacade/pkg/secretstore/gcpsecretsmanager"
	"github.com/pkg/errors"
)

func NewSecretManager(storeType secretstore.SecretStoreType) (secretstore.Interface, error) {
	switch storeType {
	case secretstore.SecretStoreTypeAzure:
		envCreds, err := azureiam.NewEnvironmentCredentials()
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		return azuresecrets.NewAzureKeyVaultSecretManager(envCreds), nil
	case secretstore.SecretStoreTypeGoogle:
		creds, err := gcpiam.DefaultCredentials()
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		return gcpsecretsmanager.NewGcpSecretsManager(*creds), nil
	}
	return nil, fmt.Errorf("unable to create manager for storeType %s", string(storeType))
}
