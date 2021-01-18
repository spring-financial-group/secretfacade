package factory

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/chrismellard/secretfacade/pkg/iam/azureiam"
	"github.com/chrismellard/secretfacade/pkg/iam/gcpiam"
	"github.com/chrismellard/secretfacade/pkg/iam/kubernetesiam"
	"github.com/chrismellard/secretfacade/pkg/iam/vaultiam"
	"github.com/chrismellard/secretfacade/pkg/secretstore"
	"github.com/chrismellard/secretfacade/pkg/secretstore/awssecretsmanager"
	"github.com/chrismellard/secretfacade/pkg/secretstore/azuresecrets"
	"github.com/chrismellard/secretfacade/pkg/secretstore/gcpsecretsmanager"
	"github.com/chrismellard/secretfacade/pkg/secretstore/kubernetessecrets"
	"github.com/chrismellard/secretfacade/pkg/secretstore/vaultsecrets"
	"github.com/pkg/errors"
)

type SecretManagerFactory struct{}

func (_ SecretManagerFactory) NewSecretManager(storeType secretstore.SecretStoreType) (secretstore.Interface, error) {
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
	case secretstore.SecretStoreTypeKubernetes:
		client, err := kubernetesiam.GetClient()
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		return kubernetessecrets.NewKubernetesSecretManager(client), nil
	case secretstore.SecretStoreTypeVault:
		creds, err := vaultiam.NewEnvironmentCreds()
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		return vaultsecrets.NewVaultSecretManager(creds.Token, creds.CaCertPath)
	case secretstore.SecretStoreTypeAws:
		sess, err := session.NewSession()
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		return awssecretsmanager.NewAwsSecretManager(sess), nil
	}
	return nil, fmt.Errorf("unable to create manager for storeType %s", string(storeType))
}
