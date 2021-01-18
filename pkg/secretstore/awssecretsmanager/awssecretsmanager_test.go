package awssecretsmanager_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/chrismellard/secretfacade/pkg/secretstore"
	"github.com/chrismellard/secretfacade/pkg/secretstore/awssecretsmanager"
	"github.com/stretchr/testify/assert"
)

func TestGetAwsSecretManager(t *testing.T) {
	s, err := session.NewSession()
	assert.NoError(t, err)
	mgr := awssecretsmanager.NewAwsSecretManager(s)
	secret, err := mgr.GetSecret("ap-southeast-2", "prod/db/creds", "")
	assert.NotEmpty(t, secret)
}

func TestSetAwsSecretManager(t *testing.T) {
	s, err := session.NewSession()
	assert.NoError(t, err)
	mgr := awssecretsmanager.NewAwsSecretManager(s)
	err = mgr.SetSecret("ap-southeast-2", "test/db/creds", &secretstore.SecretValue{Value: "supersecret"})
	assert.NoError(t, err)
}
