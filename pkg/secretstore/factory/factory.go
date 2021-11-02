package factory

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hashicorp/vault/api"
	"github.com/jenkins-x-plugins/secretfacade/pkg/iam/azureiam"
	"github.com/jenkins-x-plugins/secretfacade/pkg/iam/gcpiam"
	"github.com/jenkins-x-plugins/secretfacade/pkg/iam/kubernetesiam"
	"github.com/jenkins-x-plugins/secretfacade/pkg/iam/vaultiam"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore/awssecretsmanager"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore/awssystemmanager"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore/azuresecrets"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore/gcpsecretsmanager"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore/kubernetessecrets"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore/vaultsecrets"
	"github.com/pkg/errors"
)

type SecretManagerFactory struct{}

func (smf SecretManagerFactory) NewSecretManager(storeType secretstore.Type) (secretstore.Interface, error) {
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
		caCertPath := os.Getenv("VAULT_CACERT")
		config := api.Config{}
		err := config.ConfigureTLS(&api.TLSConfig{
			CACert: caCertPath,
		})
		if err != nil {
			return nil, errors.Wrap(err, "error configuring TLS ca cert for Hashicorp Vault API")
		}

		// ToDo: Why are we not passing the config?
		// ToDo: Change it in another PR
		client, err := api.NewClient(nil)
		if err != nil {
			return nil, errors.Wrap(err, "error creating Hashicorp Vault API client")
		}
		isExternalVault := os.Getenv("EXTERNAL_VAULT")
		if isExternalVault == "true" {
			kubeClient, err := kubernetesiam.GetClient()
			if err != nil {
				return nil, errors.Wrap(err, "error getting Kubernetes creds when attempting to create secret manager via factory")
			}
			creds, err := vaultiam.NewExternalSecretCreds(client, kubeClient)
			if err != nil {
				return nil, errors.Wrap(err, "error getting Hashicorp Vault creds when attempting to create secret manager via factory")
			}
			client.SetToken(creds.Token)
			return vaultsecrets.NewVaultSecretManager(client)
		}
		creds, err := vaultiam.NewEnvironmentCreds()
		if err != nil {
			return nil, errors.Wrap(err, "error getting Hashicorp Vault creds when attempting to create secret manager via factory")
		}

		client.SetToken(creds.Token)
		return vaultsecrets.NewVaultSecretManager(client)
	case secretstore.SecretStoreTypeAwsASM:
		sess, err := session.NewSession()
		if err != nil {
			return nil, errors.Wrap(err, "error getting AWS creds when attempting to create secret manager via factory")
		}
		return awssecretsmanager.NewAwsSecretManager(sess), nil
	case secretstore.SecretStoreTypeAwsSSM:
		sess, err := session.NewSession()
		if err != nil {
			return nil, errors.Wrap(err, "error getting AWS creds when attempting to create secret manager via factory")
		}
		return awssystemmanager.NewAwsSystemManager(sess), nil
	}
	return nil, fmt.Errorf("unable to create manager for storeType %s", string(storeType))
}
