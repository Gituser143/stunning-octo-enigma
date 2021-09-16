package client

import (
	"fmt"
	"net/http"

	graph "github.com/kiali/kiali/graph/config/cytoscape"
)

type KialiClient struct {
	httpClient *http.Client
	host       string
}

type Item struct {
	Node  *graph.NodeData   `json:"node"`
	Edges []*graph.EdgeData `json:"edges"`
}

func newItem(node *graph.NodeData) Item {
	item := Item{}

	item.Node = node

	return item
}

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
