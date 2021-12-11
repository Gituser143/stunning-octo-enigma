package kiali

import (
	"fmt"
	"net/http"

	graph "github.com/kiali/kiali/graph/config/cytoscape"
)

// Client is a type to help interact with kiali dashboards
type Client struct {
	httpClient *http.Client
	host       string
}

// Graph is a type that holds nodes and its edges indexed by node ID
type Graph map[string]*Item

// Item is a graph element
type Item struct {
	Node  *graph.NodeData   `json:"node"`
	Edges []*graph.EdgeData `json:"edges"`
}

// NewKialiClient is a constructor for type KialiClient
func NewKialiClient(host string, port int, hc *http.Client) *Client {
	kc := Client{
		host: fmt.Sprintf("%s:%d", host, port),
	}

	if hc != nil {
		kc.httpClient = hc
	} else {
		kc.httpClient = http.DefaultClient
	}
	return &kc
}
