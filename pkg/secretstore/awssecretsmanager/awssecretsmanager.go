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

func (a awsSecretsManager) SetSecret(location, secretName string, secretValue *secretstore.SecretValue) (err error) {
	// CreateSecret
	_, err = createSecret(a.session, location, secretName, *secretValue)
	if err != nil {
		// Don't return if secret already exists.
		if err.(awserr.Error).Code() != secretsmanager.ErrCodeResourceExistsException {
			return errors.Wrap(err, "error creating new secret for aws secret manager: ")
		}
	}

	// GetSecretValue + PutSecretValue/UpdateSecret
	// Get, Merge and Update
	secret, err := getExistingSecret(a.session, location, secretName)
	if err != nil {
		return errors.Wrap(err, "error retreiving existing secret for aws secret manager: ")
	}
	var existingSecretProps map[string]string
	// FIXME: If secretValue is Simple, AND then secret.SecretString is Simple.
	// getSecretPropertyMap fails
	if secretValue.Value == "" && secretValue.PropertyValues != nil {
		existingSecretProps, err = getSecretPropertyMap(secret.SecretString)
		if err != nil {
			return errors.Wrap(err, "error parsing existing secret: ")
		}
	}

	err = updateSecret(a.session, secret, secretValue.MergeExistingSecret(existingSecretProps), location, secretName)
	if err != nil {
		return errors.Wrap(err, "error updating existing secret for aws secret manager: ")
	}

	return
}

func updateSecret(session *session.Session, secret *secretsmanager.GetSecretValueOutput, newValue, location, secretName string) (err error) {
	input := &secretsmanager.PutSecretValueInput{
		SecretId:     secret.ARN,
		SecretString: aws.String(newValue),
	}
	svc := secretsmanager.New(session, aws.NewConfig().WithRegion(location))
	_, err = svc.PutSecretValue(input)
	if err != nil {
		return errors.Wrap(err, "error updating existing secret: ")
	}
	return nil
}

func getExistingSecret(session *session.Session, location, secretName string) (secret *secretsmanager.GetSecretValueOutput, err error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}
	svc := secretsmanager.New(session, aws.NewConfig().WithRegion(location))
	secret, err = svc.GetSecretValue(input)
	if err != nil {
		return
	}
	return
}

func createSecret(session *session.Session, location string, secretName string, secretValue secretstore.SecretValue) (secret *secretsmanager.GetSecretValueOutput, err error) {
	input := &secretsmanager.CreateSecretInput{
		Name:         &secretName,
		SecretString: aws.String(secretValue.ToString()),
	}
	svc := secretsmanager.New(session, aws.NewConfig().WithRegion(location))
	// svc.Config.Region = &location
	_, err = svc.CreateSecret(input)
	if err != nil {
		return
	}
	return
}

func getSecretPropertyMap(value *string) (map[string]string, error) {
	m := make(map[string]string)
	err := json.Unmarshal([]byte(*value), &m)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling AWS secrets manager secret payload in to map[string]string")
	}
	return m, nil
}
