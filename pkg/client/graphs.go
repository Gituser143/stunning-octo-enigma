package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	graph "github.com/kiali/kiali/graph/config/cytoscape"
)

// GetNamespacesGraph gives the graph for specified namespaces
func (kc *KialiClient) GetNamespacesGraph(ctx context.Context, namespaces []string) (*graph.Config, error) {

	endpoint := "kiali/api/namespaces/graph"
	namespaceStr := strings.Join(namespaces, ",")

	u := &url.URL{
		Scheme:   "http",
		Host:     kc.host,
		Path:     endpoint,
		RawQuery: fmt.Sprintf("namespaces=%s", namespaceStr),
	}

	body, err := kc.sendRequest(ctx, "GET", u.String())
	if err != nil {
		return nil, err
	}

	graphType := &graph.Config{}
	err = json.Unmarshal(body, graphType)

	return graphType, err
}

// GetWorkloadGraph gives the workload graph for a specified workload in a namespace
func (kc *KialiClient) GetWorkloadGraph(ctx context.Context, namespace, workload string) (*graph.Config, error) {
	endpoint := fmt.Sprintf("kiali/api/namespaces/%s/workloads/%s/graph", namespace, workload)

	u := &url.URL{
		Scheme: "http",
		Host:   kc.host,
		Path:   endpoint,
	}

	body, err := kc.sendRequest(ctx, "GET", u.String())
	if err != nil {
		return nil, err
	}

	graphType := &graph.Config{}
	err = json.Unmarshal(body, graphType)

	return graphType, err
}

// sendRequest constructs a request, sends it and returns the response body as a string
func (kc *KialiClient) sendRequest(ctx context.Context, method, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := kc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return body, nil
}
