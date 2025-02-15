package vaultsecrets

import (
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func NewVaultSecretManager(client *api.Client) (secretstore.Interface, error) {
	return &vaultSecretManager{client}, nil
}

type vaultSecretManager struct {
	vaultAPI *api.Client
}

func (v vaultSecretManager) GetSecret(location, secretName, secretKey string) (string, error) {
	secret, err := getSecret(v.vaultAPI, location, secretName)
	if err != nil || secret == nil {
		return "", errors.Wrapf(err, "error getting secret %s from Hasicorp vault %s", secretName, location)
	}
	mapData, err := getSecretData(secret)
	if err != nil {
		return "", errors.Wrapf(err, "error converting secret data retrieved for secret %s from Hashicorp Vault %s", secretName, location)
	}
	secretString, err := getSecretKeyString(mapData, secretKey)
	if err != nil {
		return "", errors.Wrapf(err, "error converting string data for secret %s from Hashicorp Vault %s", secretName, location)
	}
	return secretString, nil
}

func (v vaultSecretManager) SetSecret(location, secretName string, secretValue *secretstore.SecretValue) error {
	secret, err := getSecret(v.vaultAPI, location, secretName)
	if err != nil {
		return errors.Wrapf(err, "error getting secret %s in Hashicorp vault %s prior to setting", secretName, location)
	}

	newSecretData := map[string]interface{}{}
	if secret != nil && !secretValue.Overwrite {
		existingSecretData, err := getSecretData(secret)
		if err != nil {
			logrus.WithError(err).Warnf("error retrieving existing secret data in payload for secret %s in Hashicorp Vault %s", secretName, location)
		} else {
			newSecretData = existingSecretData
		}
	}

	for k, v := range secretValue.PropertyValues {
		newSecretData[k] = v
	}
	data := map[string]interface{}{
		"data": newSecretData,
	}

	_, err = v.vaultAPI.Logical().Write(secretName, data)
	if err != nil {
		return errors.Wrapf(err, "error writing secret %s to Hashicorp Vault %s", secretName, location)
	}
	return nil
}

func getSecret(client *api.Client, location, secretName string) (*api.Secret, error) {
	err := client.SetAddress(location)
	if err != nil {
		return nil, errors.Wrapf(err, "error setting location of Hashicorp vault %s on client", location)
	}
	logical := client.Logical()
	secret, err := logical.Read(secretName)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading secret %s from Hashicorp Vault API at %s", secretName, location)
	}
	return secret, nil
}

func getSecretData(secret *api.Secret) (map[string]interface{}, error) {
	data, ok := secret.Data["data"]
	if !ok {
		return nil, fmt.Errorf("data payload does not exist in Hasicorp Vault secret")
	}
	mapData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data is not of type map[string]interface{} in Hashicorp Vault secret")
	}
	return mapData, nil
}

func getSecretKeyString(secretData map[string]interface{}, secretKey string) (string, error) {
	value, ok := secretData[secretKey]
	if !ok {
		return "", fmt.Errorf("%s does not occur in secret data", secretKey)
	}
	stringValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("non string data type found in secret data for key %s", secretKey)
	}
	return stringValue, nil
}
