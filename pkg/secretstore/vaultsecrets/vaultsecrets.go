package vaultsecrets

import (
	"fmt"

	"github.com/chrismellard/secretfacade/pkg/secretstore"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

func NewVaultSecretManager(vaultToken string, caCertPath string) (secretstore.Interface, error) {
	config := api.Config{}
	err := config.ConfigureTLS(&api.TLSConfig{
		CACert: caCertPath,
	})
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	client, err := api.NewClient(nil)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	client.SetToken(vaultToken)
	return &vaultSecretManager{client}, nil
}

type vaultSecretManager struct {
	vaultApi *api.Client
}

func (v vaultSecretManager) GetSecret(location string, secretName string, secretKey string) (string, error) {
	secret, err := getSecret(v.vaultApi, location, secretName)
	if err != nil {
		return "", errors.Wrap(err, "")
	}
	mapData, err := getSecretData(secret)
	if err != nil {
		return "", errors.Wrap(err, "")
	}
	secretString, err := getSecretKeyString(mapData, secretKey)
	if err != nil {
		return "", errors.Wrap(err, "")
	}
	return secretString, nil
}

func (v vaultSecretManager) SetSecret(location string, secretName string, secretValue *secretstore.SecretValue) error {
	secret, err := getSecret(v.vaultApi, location, secretName)
	if err != nil {
		return errors.Wrap(err, "")
	}

	newSecretData := map[string]interface{}{}
	if secret != nil && !secretValue.Overwrite {
		newSecretData, err = getSecretData(secret)
		if err != nil {
			return errors.Wrap(err, "")
		}
	}

	for k, v := range secretValue.PropertyValues {
		newSecretData[k] = v
	}

	_, err = v.vaultApi.Logical().Write(fmt.Sprintf("/secret/data/%s", secretName), newSecretData)
	if err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}

func getSecret(client *api.Client, location string, secretName string) (*api.Secret, error) {
	err := client.SetAddress(location)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	logical := client.Logical()
	secret, err := logical.Read(fmt.Sprintf("/secret/data/%s", secretName))
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return secret, nil
}

func getSecretData(secret *api.Secret) (map[string]interface{}, error) {
	data, ok := secret.Data["data"]
	if !ok {
		return nil, fmt.Errorf("")
	}
	mapData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("")
	}
	return mapData, nil
}

func getSecretKeyString(secretData map[string]interface{}, secretKey string) (string, error) {
	value, ok := secretData[secretKey]
	if !ok {
		return "", fmt.Errorf("")
	}
	stringValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("")
	}
	return stringValue, nil
}
