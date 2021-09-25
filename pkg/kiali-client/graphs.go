package client

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	graph "github.com/kiali/kiali/graph/config/cytoscape"
)

// GetWorkloadGraph gives the workload graph for a specified workload in a namespace
func (kc *KialiClient) GetWorkloadGraph(ctx context.Context, namespaces []string, parameters map[string]string) (Graph, error) {
	endpoint := "kiali/api/namespaces/graph"
	parameters["namespaces"] = strings.Join(namespaces, ",")

	// Construct base URL
	u := &url.URL{
		Scheme: "http",
		Host:   kc.host,
		Path:   endpoint,
	}

	// Set query parameters
	query := u.Query()
	for parameter, value := range parameters {
		query.Add(parameter, value)
	}
	u.RawQuery = query.Encode()

	// Send request
	body, err := kc.sendRequest(ctx, "GET", u.String())
	if err != nil {
		return nil, err
	}

	graphType := &graph.Config{}
	err = json.Unmarshal(body, graphType)
	if err != nil {
		return nil, err
	}

	graph := MakeGraph(graphType)
	return graph, nil
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

// GetQueueLengths returns per workload queue lengths as a map with the
// deployment name as key and queue length as value along with the queue length
// for the 'unknown' node
func (g Graph) GetQueueLengths() (map[string]float64, float64) {
	queueLengths := make(map[string]float64)
	unkownQueueLength := 0.0

	// Iterate over items in graph. (Each item is a node along with it's edges)
	for _, item := range g {
		queueLength := 0.0

		// Iterate over an item's edges
		for _, edge := range item.Edges {
			throughput, err := strconv.ParseFloat(edge.Throughput, 64)
			if err != nil {
				throughput = 0
			}

			responseTime, err := strconv.ParseFloat(edge.ResponseTime, 64)
			if err != nil {
				responseTime = 0
			}

			// Sum up item's queue lengths as throughput * response time
			queueLength += throughput * responseTime
		}

		// Calculate item's queue length

		if item.Node.Workload == "unknown" {
			unkownQueueLength += queueLength
		} else {
			queueLengths[item.Node.Workload] = queueLength
		}
	}

	return queueLengths, unkownQueueLength
}
