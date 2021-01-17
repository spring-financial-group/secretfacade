package fake

import (
	"fmt"

	"github.com/chrismellard/secretfacade/pkg/secretstore"
)

func NewFakeSecretStore() *FakeSecretStore {
	return &FakeSecretStore{secretStores: map[string]map[string]secretType{}}
}

type FakeSecretStore struct {
	secretStores map[string]map[string]secretType
}

type secretType struct {
	secretName string
	values     secretstore.SecretValue
}

func (f FakeSecretStore) GetSecret(location string, secretName string, secretKey string) (string, error) {
	store := f.secretStores[location]
	secret := store[secretName]
	if secretKey == "" {
		return secret.values.Value, nil
	}
	for k, v := range secret.values.PropertyValues {
		if k == secretKey {
			return v, nil
		}
	}
	return "", fmt.Errorf("unable to find key %s in secret %s", secretKey, secretName)
}

func (f FakeSecretStore) SetSecret(location string, secretName string, secretValue *secretstore.SecretValue) error {
	var secrets map[string]secretType
	var ok bool
	if secrets, ok = f.secretStores[location]; !ok {
		secrets = map[string]secretType{}
		f.secretStores[location] = secrets
	}

	secrets[secretName] = secretType{
		secretName: secretName,
		values:     *secretValue,
	}

	return nil
}
