package awssecretsmanager

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/chrismellard/secretfacade/pkg/secretstore"
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
		return "", errors.Wrap(err, "")
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
	_, err := mgr.CreateSecret(input)
	if err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}
