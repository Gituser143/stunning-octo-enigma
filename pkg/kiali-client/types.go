package client

import (
	"fmt"
	"net/http"

	graph "github.com/kiali/kiali/graph/config/cytoscape"
)

// KialiClient is a type to help interact with kiali dashboards
type KialiClient struct {
	httpClient *http.Client
	host       string
}

// Item is a graph element
type Item struct {
	Node  *graph.NodeData   `json:"node"`
	Edges []*graph.EdgeData `json:"edges"`
}

func newItem(node *graph.NodeData) Item {
	item := Item{}

	item.Node = node

	return item
}

// NewKialiClient is a constructor for type KialiClient
func NewKialiClient(host string, port int, hc *http.Client) KialiClient {
	kc := KialiClient{
		host: fmt.Sprintf("%s:%d", host, port),
	}

	if hc != nil {
		kc.httpClient = hc
	} else {
		kc.httpClient = http.DefaultClient
	}
	return kc
}
