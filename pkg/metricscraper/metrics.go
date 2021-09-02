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

// LogContainerMetrics prints cpu usage of all containers in a given
// namespace along with the value of "run" label and the container name.
// Format: container_name,run_label,cpu_usage
func LogContainerMetrics(ctx context.Context, namespace string, msInterval int) error {

	clientset, err := initNewClientset()
	if err != nil {
		return err
	}

	for {
		podMetricsList, err := getContainerMetrics(ctx, namespace, clientset)
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

func initNewClientset() (*metricsv.Clientset, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}
	clientset, err := metricsv.NewForConfig(config)
	return clientset, err
}

func getContainerMetrics(ctx context.Context, namespace string, clientset *metricsv.Clientset) (*v1beta1.PodMetricsList, error) {

	podMetricsList, err := clientset.MetricsV1beta1().
		PodMetricses(namespace).
		List(ctx, metav1.ListOptions{})

	return podMetricsList, err
}
