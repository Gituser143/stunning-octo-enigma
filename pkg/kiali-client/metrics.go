package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetWorkloadMetrics gives the metrics for a specified workload in a namespace
func (kc *KialiClient) GetWorkloadMetrics(ctx context.Context, namespace, workload string) (string, error) {
	endpoint := fmt.Sprintf("kiali/api/namespaces/%s/workloads/%s/metrics", namespace, workload)

	u := &url.URL{
		Scheme: "http",
		Host:   kc.host,
		Path:   endpoint,
	}

	body, err := kc.sendRequest(ctx, "GET", u.String())
	if err != nil {
		return "", err
	}

	return string(body), err
}
