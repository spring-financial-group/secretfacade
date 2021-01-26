package kubernetessecrets

import (
	"context"
	"fmt"
	"strings"

	"github.com/chrismellard/secretfacade/pkg/secretstore"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// ReplicateToAnnotation the annotation which lists the namespaces to replicate a Secret to when using local secrets
	ReplicateToAnnotation = "secret.jenkins-x.io/replicate-to"
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

	// lets check for replicated secrets
	if secretValue.Annotations != nil {
		namespaces := secretValue.Annotations[ReplicateToAnnotation]
		if namespaces != "" {
			nsList := strings.Split(namespaces, ",")
			for _, tons := range nsList {
				err = copySecretToNamespace(k.kubeClient, tons, secret)
				if err != nil {
					return errors.Wrapf(err, "failed to replicate Secret for local backend")
				}
			}
		}
	}

	return nil
}

// copySecretToNamespace copies the given secret to the namespace
func copySecretToNamespace(kubeClient kubernetes.Interface, ns string, fromSecret *corev1.Secret) error {
	secretInterface := kubeClient.CoreV1().Secrets(ns)
	name := fromSecret.Name
	secret, err := secretInterface.Get(context.TODO(), name, metav1.GetOptions{})

	create := false
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "failed to ")
		}
		create = true
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: ns,
			},
		}
	}

	if string(fromSecret.Type) != "" {
		secret.Type = fromSecret.Type
	}
	if fromSecret.Annotations != nil {
		if secret.Annotations == nil {
			secret.Annotations = map[string]string{}
		}
		for k, v := range fromSecret.Annotations {
			secret.Annotations[k] = v
		}
	}

	if fromSecret.Labels != nil {
		if secret.Labels == nil {
			secret.Labels = map[string]string{}
		}
		for k, v := range fromSecret.Labels {
			secret.Labels[k] = v
		}
	}
	if fromSecret.Data != nil {
		if secret.Data == nil {
			secret.Data = map[string][]byte{}
		}
		for k, v := range fromSecret.Data {
			secret.Data[k] = v
		}
	}

	if create {
		_, err = secretInterface.Create(context.TODO(), secret, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to create Secret %s in namespace %s", name, ns)
		}
		//fmt.Printf("created replica Secret %s in namespace %s\n", name, ns)
		return nil
	}
	_, err = secretInterface.Update(context.TODO(), secret, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to update Secret %s in namespace %s", name, ns)
	}
	//fmt.Printf("updated replica Secret %s in namespace %s\n", name, ns)
	return nil
}
