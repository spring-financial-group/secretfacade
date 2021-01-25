package kubernetessecrets

import (
	"context"
	"fmt"

	"github.com/chrismellard/secretfacade/pkg/secretstore"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func NewKubernetesSecretManager(kubeClient kubernetes.Interface) secretstore.Interface {
	return &kubernetesSecretManager{kubeClient: kubeClient}
}

type kubernetesSecretManager struct {
	kubeClient kubernetes.Interface
}

func (k kubernetesSecretManager) GetSecret(namespace string, secretName string, secretKey string) (string, error) {
	secret, err := k.kubeClient.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return "", errors.Wrapf(err, "failed to get secret %s from namespace %s", secretName, namespace)
	}
	secretData, ok := secret.Data[secretKey]
	if ok {
		return string(secretData), nil
	}
	secretString, ok := secret.StringData[secretKey]
	if ok {
		return secretString, nil
	}
	return "", fmt.Errorf("failed to get secret %s from namespace %s", secretName, namespace)
}

func (k kubernetesSecretManager) SetSecret(namespace string, secretName string, secretValue *secretstore.SecretValue) error {
	create := false
	secretInterface := k.kubeClient.CoreV1().Secrets(namespace)
	secret, err := secretInterface.Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "failed to ")
		}
		create = true
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: namespace,
			},
			Type: corev1.SecretTypeOpaque,
		}
	}

	secret.Type = corev1.SecretTypeOpaque
	if string(secretValue.SecretType) != "" {
		secret.Type = secretValue.SecretType
	}
	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	for k, v := range secretValue.PropertyValues {
		secret.Data[k] = []byte(v)
	}

	if secretValue.Labels != nil {
		if secret.Labels == nil {
			secret.Labels = map[string]string{}
		}
		for k, v := range secretValue.Labels {
			secret.Labels[k] = v
		}
	}
	if secretValue.Annotations != nil {
		if secret.Annotations == nil {
			secret.Annotations = map[string]string{}
		}
		for k, v := range secretValue.Annotations {
			secret.Annotations[k] = v
		}
	}

	if create {
		_, err = secretInterface.Create(context.TODO(), secret, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to create Secret %s in namespace %s", secretName, namespace)
		}
	} else {
		_, err = secretInterface.Update(context.TODO(), secret, metav1.UpdateOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to update Secret %s in namespace %s", secretName, namespace)
		}
	}
	return nil
}
