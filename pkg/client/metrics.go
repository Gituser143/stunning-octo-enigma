package client

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

func GetMetrics(namespace string) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
		}
	}
	clientset, err := metricsv.NewForConfig(config)
	podMetricsList, err := clientset.MetricsV1beta1().
		PodMetricses(namespace).
		List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		log.Fatal((err))
	}

	fmt.Println(podMetricsList)
}
