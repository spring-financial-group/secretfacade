package awssecretsmanager

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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

func (a awsSecretsManager) SetSecret(location string, secretName string, secretValue *secretstore.SecretValue) (err error) {
	secret, err := existingSecret(a.session, &location, &secretName)
	if err != nil {
		return errors.Wrap(err, "error checking for existing secret:")
	}
	if secret != nil {
		// Get, Merge and Update
		input := &secretsmanager.GetSecretValueInput{
			SecretId: secret.ARN,
		}
		mgr := secretsmanager.New(a.session, aws.NewConfig().WithRegion(location))
		// mgr.Config.Region = &location
		_, _ = mgr.GetSecretValue(input)
		return
	}

	if secretValue.Value == "" && secretValue.PropertyValues != nil {
		str, _ := json.Marshal(secretValue.PropertyValues)
		secretValue.Value = string(str)
	}
	input := &secretsmanager.CreateSecretInput{
		Name:         &secretName,
		SecretString: &secretValue.Value,
	}
	svc := secretsmanager.New(a.session, aws.NewConfig().WithRegion(location))
	// mgr.Config.Region = &location
	_, err = svc.CreateSecret(input)
	if err != nil {
		return errors.Wrap(err, "error setting secret for aws secret manager: ")
	}
	return
}

func existingSecret(session *session.Session, location, secretName *string) (secret *secretsmanager.SecretListEntry, err error) {
	nameFilter := &secretsmanager.Filter{
		Key:    aws.String(secretsmanager.FilterNameStringTypeName),
		Values: []*string{secretName},
	}
	input := &secretsmanager.ListSecretsInput{
		Filters:    []*secretsmanager.Filter{nameFilter},
		MaxResults: aws.Int64(int64(1)),
	}

	svc := secretsmanager.New(session, aws.NewConfig().WithRegion(*location))
	secrets, err := svc.ListSecrets(input)
	if err != nil {
		return
	}
	secret = secrets.SecretList[0]
	return
}
