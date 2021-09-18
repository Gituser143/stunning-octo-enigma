package metricscraper

import (
	"context"
	"log"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// MetricClient is a wrapper around a clientset.
//
// NOTE: the *minimum* scrape interval of the metric-server
// itself is 10 seconds. Even if we provide frequencies of
// less than 10 seconds in the below functions, its highly
// likely that we are going to end up recieving more or less
// similar values. Because of this, it would be safe to assume
// that any and all operations that make use of these metrics
// are long running operations.
type MetricClient struct {
	client *metricsv.Clientset
}

// NewMetricClient inits a new clientset from a local kubeconfig and slaps a MetricClient around it.
func NewMetricClient() (*MetricClient, error) {
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

	return &MetricClient{clientset}, err
}

// StreamAllPodMetrics returns a channel that has a []v1beta1.PodMetrics put in it every freq Duration.
func (c *MetricClient) StreamAllPodMetrics(ctx context.Context, namespace string, freq time.Duration) <-chan []v1beta1.PodMetrics {
	metricChan := make(chan []v1beta1.PodMetrics, 1)
	go func() {
		ticker := time.NewTicker(freq)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return // prevent goroutine leak.
			case <-ticker.C:
				metrics, err := c.GetAllPodMetrics(ctx, namespace)
				// hack. just log error, don't really do anything about it for now.
				if err != nil {
					log.Println(err)
				}
				metricChan <- metrics
			}
		}
	}()

	return metricChan
}

// TODO: stream pod metric, map deployment to metric type.

// StreamPodMetrics returns a channel that has *v1beta1.PodMetrics put in it every freq Duration.
func (c *MetricClient) StreamPodMetrics(ctx context.Context, namespace, name string, freq time.Duration) <-chan *v1beta1.PodMetrics {
	metricChan := make(chan *v1beta1.PodMetrics, 1)
	go func() {
		ticker := time.NewTicker(freq)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return // prevent goroutine leak.
			case <-ticker.C:
				metrics, err := c.GetPodMetrics(ctx, namespace, name)
				// hack. just log error, don't really do anything about it for now.
				if err != nil {
					log.Println(err)
				}
				metricChan <- metrics
			}
		}
	}()

	return metricChan
}

// GetAllPodMetrics returns metrics of all pods in a namespace.
func (c *MetricClient) GetAllPodMetrics(ctx context.Context, namespace string) ([]v1beta1.PodMetrics, error) {
	podMetrices, err := c.client.MetricsV1beta1().
		PodMetricses(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return podMetrices.Items, nil
}

// GetPodMetrics gets the metrics of a named pod in a particular namespace.
func (c *MetricClient) GetPodMetrics(ctx context.Context, namespace, name string) (*v1beta1.PodMetrics, error) {
	return c.client.MetricsV1beta1().PodMetricses(namespace).Get(ctx, name, metav1.GetOptions{})
}
