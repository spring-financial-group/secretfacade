package awssecretsmanager

import (
	"encoding/json"

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

func (a awsSecretsManager) SetSecret(location, secretName string, secretValue *secretstore.SecretValue) (err error) {

	err = updateExistingSecret(a.session, *secretValue, &location, &secretName)
	if err != nil {
		return errors.Wrap(err, "error updating existing secret for aws secret manager: ")
	}

	input := &secretsmanager.CreateSecretInput{
		Name:         &secretName,
		SecretString: aws.String(secretValue.ToString()),
	}
	svc := secretsmanager.New(a.session, aws.NewConfig().WithRegion(location))
	// mgr.Config.Region = &location
	_, err = svc.CreateSecret(input)
	if err != nil {
		return errors.Wrap(err, "error setting secret for aws secret manager: ")
	}
	return
}

func updateExistingSecret(session *session.Session, sv secretstore.SecretValue, location, secretName *string) (err error) {
	_, err = existingSecret(session, location, secretName)
	if err != nil {
		return errors.Wrap(err, "error searching for existing secret: ")
	}
	// Get, Merge and Update
	var existingSecretProps map[string]string
	secret, err := getExistingSecret(session, *location, *secretName)
	if err != nil {
		return errors.Wrap(err, "error retreiving existing secret: ")
	}
	// FIXME: If secretValue is Simple, AND then secret.SecretString is Simple.
	// getSecretPropertyMap fails
	if sv.Value == "" && sv.PropertyValues != nil {
		existingSecretProps, err = getSecretPropertyMap(secret.SecretString)
		if err != nil {
			return errors.Wrap(err, "error parsing existing secret: ")
		}
	}

	input := &secretsmanager.UpdateSecretInput{
		SecretId:     secret.ARN,
		SecretString: aws.String(sv.MergeExistingSecret(existingSecretProps)),
	}
	svc := secretsmanager.New(session, aws.NewConfig().WithRegion(*location))
	_, err = svc.UpdateSecret(input)
	if err != nil {
		return errors.Wrap(err, "error updating existing secret: ")
	}
	return nil
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
	if len(secrets.SecretList) == 0 {
		err = errors.New("no existing secrets")
		return
	}
	secret = secrets.SecretList[0]
	return
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

func getSecretPropertyMap(value *string) (map[string]string, error) {
	m := make(map[string]string)
	err := json.Unmarshal([]byte(*value), &m)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling AWS secrets manager secret payload in to map[string]string")
	}
	return m, nil
}
