package kubernetesiam

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func InClusterClient() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "error getting in cluster config for k8s")
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "error creating clientset from in cluster config for k8s")
	}
	return clientset, nil
}

func OutOfClusterClient() (kubernetes.Interface, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "error getting config for out of cluster k8s")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "error creating clientset for out of cluster k8s")
	}
	return clientset, nil
}

func GetClient() (kubernetes.Interface, error) {
	inCluster, err := InClusterClient()
	if err == nil {
		return inCluster, nil
	}

	outOfCluster, err := OutOfClusterClient()
	if err == nil {
		return outOfCluster, nil
	}

	return nil, fmt.Errorf("unable to configure kubernetes client")
}
