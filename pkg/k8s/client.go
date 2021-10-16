package k8s

import (
	"context"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// K8sClient is a wrapper around a clientset.
type K8sClient struct {
	client *kubernetes.Clientset
}

// NewK8sClient inits a new clientset from a local kubeconfig and slaps a K8sClient around it.
func NewK8sClient() (*K8sClient, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &K8sClient{client}, nil
}

// GetDeploymentNames gets only names of all deployments.
func (c *K8sClient) GetDeploymentNames(ctx context.Context, namespace string) ([]string, error) {
	deployments, err := c.GetDeployments(ctx, namespace)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(deployments))
	for i, d := range deployments {
		names[i] = d.Name
	}

	return names, nil
}

// GetDeployments lists out all deployments and returns a slice of it.
func (c *K8sClient) GetDeployments(ctx context.Context, namespace string) ([]appsv1.Deployment, error) {
	depList, err := c.client.AppsV1().
		Deployments(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return depList.Items, nil
}

// GetPodNames blah blah.
func (c *K8sClient) GetPodNames(ctx context.Context, namespace string) ([]string, error) {
	pods, err := c.GetPods(ctx, namespace)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(pods))
	for i, d := range pods {
		names[i] = d.Name
	}

	return names, nil
}

// GetPods gets blah blah.
func (c *K8sClient) GetPods(ctx context.Context, namespace string) ([]apiv1.Pod, error) {
	podList, err := c.client.CoreV1().
		Pods(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return podList.Items, nil
}

// GetServiceNames blah blah.
func (c *K8sClient) GetServiceNames(ctx context.Context, namespace string) ([]string, error) {
	svcs, err := c.GetServices(ctx, namespace)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(svcs))
	for i, d := range svcs {
		names[i] = d.Name
	}

	return names, nil
}

// GetServices gets blah blah.
func (c *K8sClient) GetServices(ctx context.Context, namespace string) ([]apiv1.Service, error) {
	svcs, err := c.client.CoreV1().
		Services(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return svcs.Items, nil
}

// ScaleDeployment pulls one scale scene on a deployment.
func (c *K8sClient) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error {
	s, err := c.client.AppsV1().
		Deployments(namespace).
		GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	sc := s.DeepCopy()
	sc.Spec.Replicas = replicas

	_, err = c.client.AppsV1().
		Deployments(namespace).
		UpdateScale(ctx, name, sc, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

// Gets current Replica Count of given deployment in a namespace
func (c *K8sClient) GetCurrentReplicaCount(ctx context.Context, namespace, name string) (int32, error) {
	s, err := c.client.AppsV1().
		Deployments(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return 0, err
	}

	return s.Status.Replicas, err
}
