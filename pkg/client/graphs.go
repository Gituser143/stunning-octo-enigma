package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// GetNamespacesGraph gives the graph for specified namespaces
// TODO: Should return a variable of type graph
func (kc *KialiClient) GetNamespacesGraph(ctx context.Context, namespaces []string) error {

	endpoint := "/api/namespaces/graph"
	namespaceStr := strings.Join(namespaces, ",")
	url := fmt.Sprintf("http://%s:%d/kiali%s?namespaces=%s", kc.host, kc.port, endpoint, namespaceStr)

	body, err := kc.sendRequest(ctx, "GET", url)
	if err != nil {
		return err
	}

	fmt.Println(body)

	return nil
}

func (kc *KialiClient) GetWorkloadGraph(ctx context.Context, namespace, workload string) error {

	endpoint := fmt.Sprintf("/api/namespaces/%s/workloads/%s/graph", namespace, workload)
	url := fmt.Sprintf("http://%s:%d/kiali%s", kc.host, kc.port, endpoint)

	body, err := kc.sendRequest(ctx, "GET", url)
	if err != nil {
		return err
	}

	fmt.Println(body)

	return nil
}

// sendRequest constructs a request, sends it and returns the response body as a string
func (kc *KialiClient) sendRequest(ctx context.Context, method, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := kc.httpClient.Do(req)
	if err != nil {
		return "", err
	}

	bodyBs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body := string(bodyBs)
	return body, nil
}
