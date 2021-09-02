package client

import (
	"context"
	"log"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func Scale(depName, depNamespace string) error {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	s, err := client.AppsV1().
		Deployments(depNamespace).
		GetScale(context.TODO(), depName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	sc := *s
	sc.Spec.Replicas = 2

	us, err := client.AppsV1().
		Deployments(depNamespace).
		UpdateScale(context.TODO(),
			depName, &sc, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	log.Println(*us)
	return nil
}
