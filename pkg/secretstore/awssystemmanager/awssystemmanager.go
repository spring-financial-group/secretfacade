package awssystemmanager

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore"
	"github.com/pkg/errors"
)

func NewAwsSystemManager(session *session.Session) secretstore.Interface {
	return awsSystemManager{session}
}

type awsSystemManager struct {
	session *session.Session
}

func (a awsSystemManager) GetSecret(location string, secretName string, _ string) (string, error) {
	input := &ssm.GetParameterInput{
		Name:           aws.String(secretName),
	}
	mgr := ssm.New(a.session, aws.NewConfig().WithRegion(location))
	mgr.Config.Region = &location
	result, err := mgr.GetParameter(input)
	if err != nil {
		return "", errors.Wrap(err, "error retrieving secret from aws parameter store")
	}
	return result.String(), nil
}

func (a awsSystemManager) SetSecret(location string, secretName string, secretValue *secretstore.SecretValue) error {
	input := &ssm.PutParameterInput{
		Name:           &secretName,
		Value:          &secretValue.Value,
	}
	mgr := ssm.New(a.session, aws.NewConfig().WithRegion(location))
	mgr.Config.Region = &location

	_, err := mgr.PutParameter(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ssm.ErrCodeParameterAlreadyExists:
				return errors.Wrap(err, "Secret Already Exists")
			}
			return errors.Wrap(err, "error setting secret for aws parameter store")
	}
	}
	return nil
}
