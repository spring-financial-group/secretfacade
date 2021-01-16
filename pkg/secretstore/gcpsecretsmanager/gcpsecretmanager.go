package gcpsecretsmanager

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/chrismellard/secretfacade/pkg/iam/gcp"
	"github.com/chrismellard/secretfacade/pkg/secretstore"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

type GcpSecretsManagerOperations struct {
}

func (g *GcpSecretsManagerOperations) SetSecret(projectId string, secretName string, secretValue *secretstore.SecretValue) error {

	client, closer, err := getSecretOpsClient()
	if err != nil {
		return errors.Wrap(err, "")
	}
	defer closer()

	secret, err := getSecret(client, projectId, secretName)
	if err != nil {
		secret, err = createSecret(client, projectId, secretName)
		if err != nil {
			return errors.Wrap(err, "")
		}
	}

	req := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(secretValue.ToString()),
		},
	}
	_, err = client.AddSecretVersion(context.TODO(), req)
	if err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}

func (_ *GcpSecretsManagerOperations) GetSecret(projectId string, secretName string, _ string) (string, error) {
	client, closer, err := getSecretOpsClient()
	if err != nil {
		return "", errors.Wrap(err, "")
	}
	defer closer()

	secret, err := getSecret(client, projectId, secretName)
	if err != nil {
		return "", errors.Wrap(err, "")
	}
	return secret.String(), nil
}

func getSecretOpsClient() (*secretmanager.Client, func(), error) {
	creds, err := gcp.DefaultCredentials(context.TODO())
	if err != nil {
		return nil, nil, errors.Wrap(err, "")
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

func createSecret(client *secretmanager.Client, projectId string, secretName string) (*secretmanagerpb.Secret, error) {
	req := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", projectId),
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
		return nil, errors.Wrap(err, "")
	}
	return secret, nil
}

func getSecret(client *secretmanager.Client, projectId string, secretName string) (*secretmanagerpb.Secret, error) {

	req := &secretmanagerpb.GetSecretRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s", projectId, secretName),
	}
	secret, err := client.GetSecret(context.TODO(), req)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return secret, nil
}
