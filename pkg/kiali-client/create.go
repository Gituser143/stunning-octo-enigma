package client

import (
	"encoding/json"
	"fmt"

	graph "github.com/kiali/kiali/graph/config/cytoscape"
)

func MakeGraph(graphType *graph.Config) (map[string]*Item, error) {
	adjList := make(map[string]*Item)
	for _, node := range graphType.Elements.Nodes {
		item := newItem(node.Data)
		adjList[node.Data.ID] = &item
	}
	for _, edge := range graphType.Elements.Edges {
		if adjList[edge.Data.Source].Edges != nil {
			adjList[edge.Data.Source].Edges = append(adjList[edge.Data.Source].Edges, edge.Data)
		} else {
			adjList[edge.Data.Source].Edges = make([]*graph.EdgeData, 0)
			adjList[edge.Data.Source].Edges = append(adjList[edge.Data.Source].Edges, edge.Data)
		}
	}

	jsonAdjList, err := json.Marshal(adjList)
	if err == nil {
		fmt.Printf(string(jsonAdjList))
	} else {
		fmt.Printf(err.Error())
	}
	return adjList, err
}
