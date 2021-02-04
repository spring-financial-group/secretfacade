package factory

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jenkins-x-plugins/secretfacade/pkg/iam/azureiam"
	"github.com/jenkins-x-plugins/secretfacade/pkg/iam/gcpiam"
	"github.com/jenkins-x-plugins/secretfacade/pkg/iam/kubernetesiam"
	"github.com/jenkins-x-plugins/secretfacade/pkg/iam/vaultiam"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore/awssecretsmanager"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore/azuresecrets"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore/gcpsecretsmanager"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore/kubernetessecrets"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore/vaultsecrets"
	"github.com/pkg/errors"
)

type SecretManagerFactory struct{}

func (_ SecretManagerFactory) NewSecretManager(storeType secretstore.SecretStoreType) (secretstore.Interface, error) {
	switch storeType {
	case secretstore.SecretStoreTypeAzure:
		envCreds, err := azureiam.NewEnvironmentCredentials()
		if err != nil {
			return nil, errors.Wrap(err, "error getting azure creds when attempting to create secret manager via factory")
		}
		return azuresecrets.NewAzureKeyVaultSecretManager(envCreds), nil
	case secretstore.SecretStoreTypeGoogle:
		creds, err := gcpiam.DefaultCredentials()
		if err != nil {
			return nil, errors.Wrap(err, "error getting Google creds when attempting to create secret manager via factory")
		}
		return gcpsecretsmanager.NewGcpSecretsManager(*creds), nil
	case secretstore.SecretStoreTypeKubernetes:
		client, err := kubernetesiam.GetClient()
		if err != nil {
			return nil, errors.Wrap(err, "error getting Kubernetes creds when attempting to create secret manager via factory")
		}
		return kubernetessecrets.NewKubernetesSecretManager(client), nil
	case secretstore.SecretStoreTypeVault:
		creds, err := vaultiam.NewEnvironmentCreds()
		if err != nil {
			return nil, errors.Wrap(err, "error getting Hashicorp Vault creds when attempting to create secret manager via factory")
		}
		return vaultsecrets.NewVaultSecretManager(creds.Token, creds.CaCertPath)
	case secretstore.SecretStoreTypeAws:
		sess, err := session.NewSession()
		if err != nil {
			return nil, errors.Wrap(err, "error getting AWS creds when attempting to create secret manager via factory")
		}
		return awssecretsmanager.NewAwsSecretManager(sess), nil
	}
	return nil, fmt.Errorf("unable to create manager for storeType %s", string(storeType))
}
