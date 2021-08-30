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
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := kc.httpClient.Do(req)
	if err != nil {
		return err
	}

	bodyBs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body := string(bodyBs)
	fmt.Println(body)

	return nil
}
