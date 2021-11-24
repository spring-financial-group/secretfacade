package gcpsecretsmanager

import (
	"context"
	"encoding/json"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/jenkins-x-plugins/secretfacade/pkg/iam/gcpiam"
	"github.com/jenkins-x-plugins/secretfacade/pkg/secretstore"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

func NewGcpSecretsManager(creds google.Credentials) secretstore.Interface {
	return &gcpSecretsManager{creds}
}

type gcpSecretsManager struct {
	creds google.Credentials
}

func (g *gcpSecretsManager) SetSecret(projectID, secretName string, secretValue *secretstore.SecretValue) error {

	client, closer, err := getSecretOpsClient()
	if err != nil {
		return errors.Wrapf(err, "error setting GCP Secrets Manager secret %s in project %s", secretName, projectID)
	}
	defer closer()

	var existingSecretProps map[string]string
	secret, err := getSecret(client, projectID, secretName)
	if err != nil {
		secret, err = createSecret(client, projectID, secretName)
		if err != nil {
			return errors.Wrapf(err, "error creating new secret %s in GCP secret manager project %s", secretName, projectID)
		}
	} else if secretValue.Value == "" && secretValue.PropertyValues != nil {
		sv, err := getSecretValue(client, projectID, secretName)
		if err != nil {
			return errors.Wrapf(err, "error getting GCP secrets manager secret value for secret name %s in project %s", secretName, projectID)
		}
		existingSecretProps, err = getSecretPropertyMap(sv)
		if err != nil {
			return errors.Wrap(err, "error getting secret property map")
		}
	}

	req := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(secretValue.MergeExistingSecret(existingSecretProps)),
		},
	}
	_, err = client.AddSecretVersion(context.TODO(), req)
	if err != nil {
		return errors.Wrapf(err, "unable to set secret %s in GCP secret manager project %s", secretName, projectID)
	}
	return nil
}

func (g *gcpSecretsManager) GetSecret(projectID, secretName, secretKey string) (string, error) {
	client, closer, err := getSecretOpsClient()
	if err != nil {
		return "", errors.Wrap(err, "error creating GCP secret manager client")
	}
	defer closer()

	secret, err := getSecretValue(client, projectID, secretName)
	if err != nil {
		return "", errors.Wrapf(err, "error getting secret %s for GCP secret manager in project %s", secretName, projectID)
	}
	var secretString string
	if secretKey != "" {
		secretString, err = getSecretProperty(secret, secretKey)
		if err != nil {
			return "", errors.Wrapf(err, "error retrieving secret property from secret %s returned from GCP secrets manager in project %s", secretName, projectID)
		}
	} else {
		secretString = string(secret.Data)
	}
	return secretString, nil
}

func getSecretPropertyMap(v *secretmanagerpb.SecretPayload) (map[string]string, error) {
	m := make(map[string]string)
	err := json.Unmarshal(v.Data, &m)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling GCP secrets manager secret payload in to map[string]string")
	}
	return m, nil
}

func getSecretProperty(v *secretmanagerpb.SecretPayload, propertyName string) (string, error) {
	m, err := getSecretPropertyMap(v)
	if err != nil {
		return "", errors.Wrapf(err, "error reading property %s from secret JSON object", propertyName)
	}
	return m[propertyName], nil
}

func getSecretOpsClient() (*secretmanager.Client, func(), error) {
	creds, err := gcpiam.DefaultCredentials()
	if err != nil {
		return nil, nil, errors.Wrap(err, "error getting GCP default credentials")
	}
	client, err := secretmanager.NewClient(context.TODO(),
		option.WithGRPCDialOption(
			grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
		),
		option.WithTokenSource(oauth.TokenSource{TokenSource: creds.TokenSource}),
	)
	if err != nil {
		return nil, nil, err
	}

	return client, func() { _ = client.Close() }, nil
}

func createSecret(client *secretmanager.Client, projectID, secretName string) (*secretmanagerpb.Secret, error) {
	req := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", projectID),
		SecretId: secretName,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}
	secret, err := client.CreateSecret(context.TODO(), req)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating secret %s in GCP secrets manager for project %s", secretName, projectID)
	}
	return secret, nil
}

func getSecret(client *secretmanager.Client, projectID, secretName string) (*secretmanagerpb.Secret, error) {

	req := &secretmanagerpb.GetSecretRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s", projectID, secretName),
	}
	secret, err := client.GetSecret(context.TODO(), req)

	if err != nil {
		return nil, errors.Wrapf(err, "error getting secret %s for GCP secrets manager project %s", secretName, projectID)
	}
	return secret, nil
}

func getSecretValue(client *secretmanager.Client, projectID, secretName string) (*secretmanagerpb.SecretPayload, error) {

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, secretName),
	}
	secret, err := client.AccessSecretVersion(context.TODO(), req)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting secret value for secret %s for GCP secrets manager project %s", secretName, projectID)
	}
	return secret.Payload, nil
}
