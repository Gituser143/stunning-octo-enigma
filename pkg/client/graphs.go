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
func (kc *KialiClient) GetNamespacesGraph(ctx context.Context, host string, port int, namespaces []string) error {
	client := kc.HttpClient

	endpoint := "/api/namespaces/graph"
	namespaceStr := strings.Join(namespaces, ",")
	url := fmt.Sprintf("http://%s:%d/kiali%s?namespaces=%s", host, port, endpoint, namespaceStr)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	bodyBs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	body := string(bodyBs)
	fmt.Println(body)

	return nil
}
