package vaultsecrets

import (
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore"
	"github.com/pkg/errors"
)

func NewVaultSecretManager(client *api.Client) (secretstore.Interface, error) {
	return &vaultSecretManager{client}, nil
}

type vaultSecretManager struct {
	vaultApi *api.Client
}

func (v vaultSecretManager) GetSecret(location string, secretName string, secretKey string) (string, error) {
	secret, err := getSecret(v.vaultApi, location, secretName)
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

func (v vaultSecretManager) SetSecret(location string, secretName string, secretValue *secretstore.SecretValue) error {
	secret, err := getSecret(v.vaultApi, location, secretName)
	if err != nil {
		return errors.Wrapf(err, "error getting secret %s in Hashicorp vault %s prior to setting", secretName, location)
	}

	newSecretData := map[string]interface{}{}
	if secret != nil && !secretValue.Overwrite {
		newSecretData, err = getSecretData(secret)
		if err != nil {
			return errors.Wrapf(err, "error retrieving secret data in payload for secret %s in Hashicorp Vault %s", secretName, location)
		}
	}

	for k, v := range secretValue.PropertyValues {
		newSecretData[k] = v
	}
	data := map[string]interface{}{
		"data": newSecretData,
	}

	_, err = v.vaultApi.Logical().Write(secretName, data)
	if err != nil {
		return errors.Wrapf(err, "error writing secret %s to Hashicorp Vault %s", secretName, location)
	}
	return nil
}

func getSecret(client *api.Client, location string, secretName string) (*api.Secret, error) {
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
