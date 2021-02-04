package awssecretsmanager

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore"
	"github.com/pkg/errors"
)

func NewAwsSecretManager(session *session.Session) secretstore.Interface {
	return awsSecretsManager{session}
}

type awsSecretsManager struct {
	session *session.Session
}

func (a awsSecretsManager) GetSecret(location string, secretName string, _ string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}
	mgr := secretsmanager.New(a.session, aws.NewConfig().WithRegion(location))
	mgr.Config.Region = &location
	result, err := mgr.GetSecretValue(input)
	if err != nil {
		return "", errors.Wrap(err, "error retrieving secret from aws secret manager")
	}
	return *result.SecretString, nil
}

func (a awsSecretsManager) SetSecret(location string, secretName string, secretValue *secretstore.SecretValue) error {
	input := &secretsmanager.CreateSecretInput{
		Name:         &secretName,
		SecretString: &secretValue.Value,
	}
	mgr := secretsmanager.New(a.session, aws.NewConfig().WithRegion(location))
	mgr.Config.Region = &location
	// TODO - Put a fix in here to detect if secret already exists
	_, err := mgr.CreateSecret(input)
	if err != nil {
		return errors.Wrap(err, "error setting secret for aws secret manager")
	}
	return nil
}
