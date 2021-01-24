package azuresecrets_test

import (
	"os"
	"testing"

	"github.com/chrismellard/secretfacade/pkg/iam/azureiam"
	"github.com/chrismellard/secretfacade/pkg/secretstore"
	"github.com/chrismellard/secretfacade/pkg/secretstore/azuresecrets"
	"github.com/stretchr/testify/assert"
)

func TestAzureKey(t *testing.T) {

	keyVaultName := os.Getenv("AZURE_KEY_VAULT")
	creds, err := azureiam.NewEnvironmentCredentials()
	assert.NoError(t, err)
	secretMgr := azuresecrets.NewAzureKeyVaultSecretManager(creds)

	err = secretMgr.SetSecret(keyVaultName, "testsecret", &secretstore.SecretValue{
		PropertyValues: map[string]string{
			"username": "thisisausername",
			"password": "thisisapassword",
		},
	})
	assert.NoError(t, err)

	username, err := secretMgr.GetSecret(keyVaultName, "testsecret", "username")
	assert.NoError(t, err)
	assert.Equal(t, "thisisausername", username)

}
