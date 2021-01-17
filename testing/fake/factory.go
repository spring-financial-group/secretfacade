package fake

import "github.com/chrismellard/secretfacade/pkg/secretstore"

type FakeSecretManagerFactory struct{}

func (_ FakeSecretManagerFactory) NewSecretManager(_ secretstore.SecretStoreType) (secretstore.Interface, error) {
	return NewFakeSecretStore(), nil
}
