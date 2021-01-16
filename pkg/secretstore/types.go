package secretstore

// SecretStoreType describes a secrets backend
type SecretStoreType string

const (
	// SecretStoreTypeAzure Azure Key Vault as the secret store
	SecretStoreTypeAzure      SecretStoreType = "azureKeyVault"
	SecretStoreTypeGoogle     SecretStoreType = "gcpSecretsManager"
	SecretStoreTypeKubernetes SecretStoreType = "kubernetes"
)
