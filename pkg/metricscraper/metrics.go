package client

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

func LogContainerMetrics(ctx context.Context, namespace string, msInterval int) error {

	for {
		podMetricsList, err := getContainerMetrics(ctx, namespace)
		if err != nil {
			return err
		}

		for _, item := range podMetricsList.Items {
			for _, container := range item.Containers {
				fmt.Printf(
					"%s,%s,%s\n",
					item.Labels["run"],
					container.Name,
					container.Usage.Cpu(),
				)
			}
		}
		time.Sleep(time.Duration(msInterval) * time.Millisecond)
	}
}

func getContainerMetrics(ctx context.Context, namespace string) (*v1beta1.PodMetricsList, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}
	clientset, err := metricsv.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	podMetricsList, err := clientset.MetricsV1beta1().
		PodMetricses(namespace).
		List(ctx, metav1.ListOptions{})

	return podMetricsList, err
}
