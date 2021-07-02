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

func (a awsSecretsManager) SetSecret(location string, secretName string, secretValue *secretstore.SecretValue) error {
	if secretExists(a.session, &location, &secretName) {
		return nil
	} else {
		if secretValue.Value == "" && secretValue.PropertyValues != nil {
			str, _ := json.Marshal(secretValue.PropertyValues)
			secretValue.Value = string(str)
		}
		input := &secretsmanager.CreateSecretInput{
			Name:               &secretName,
			SecretString:       &secretValue.Value,
		}
		mgr := secretsmanager.New(a.session, aws.NewConfig().WithRegion(location))
		// mgr.Config.Region = &location
		_, err := mgr.CreateSecret(input)
		if err != nil {
			aerr := err.(awserr.Error)
			return errors.Wrap(aerr.OrigErr(), "error setting secret for aws secret manager: " + aerr.Error())
		}
		return nil
	}
}

func secretExists(session *session.Session, location, secretName *string) bool {
	// ListSecret
	maxResults := int64(1)
	nameFilterKey := secretsmanager.FilterNameStringTypeName
	nameFilter := &secretsmanager.Filter{
		Key:    &nameFilterKey,
		Values: []*string{secretName},
	}
	input := &secretsmanager.ListSecretsInput{
		Filters:    []*secretsmanager.Filter{nameFilter},
		MaxResults: &maxResults,
	}

	mgr := secretsmanager.New(session, aws.NewConfig().WithRegion(*location))
	secrets, err := mgr.ListSecrets(input)
	if err == nil {
		return len(secrets.SecretList) >= 1
	}
	return false
}
