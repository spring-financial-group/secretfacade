package secretstore

type Interface interface {
	GetSecret(location string, secretName string, secretKey string) (string, error)
	SetSecret(location string, secretName string, secretValue *SecretValue) error
}

type FactoryInterface interface {
	NewSecretManager(storeType SecretStoreType) (Interface, error)
}
