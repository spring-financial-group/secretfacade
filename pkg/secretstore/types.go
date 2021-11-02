package secretstore

// SecretStoreType describes a secrets backend
type Type string

const (
	// SecretStoreTypeAzure Azure Key Vault as the secret store
	SecretStoreTypeAzure      Type = "azureKeyVault"
	SecretStoreTypeGoogle     Type = "gcpSecretsManager"
	SecretStoreTypeKubernetes Type = "kubernetes"
	SecretStoreTypeVault      Type = "vault"
	SecretStoreTypeAwsASM     Type = "secretsManager"
	SecretStoreTypeAwsSSM     Type = "systemManager"
)
