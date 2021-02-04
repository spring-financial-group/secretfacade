package fake

import "github.com/jenkins-x-plugins/secretfacade/pkg/secretstore"

type FakeSecretManagerFactory struct {
	secretStore *FakeSecretStore
}

func (f *FakeSecretManagerFactory) NewSecretManager(_ secretstore.SecretStoreType) (secretstore.Interface, error) {
	if f.secretStore == nil {
		f.secretStore = NewFakeSecretStore()
	}
	return f.secretStore, nil
}

func (f FakeSecretManagerFactory) GetSecretStore() *FakeSecretStore {
	return f.secretStore
}
