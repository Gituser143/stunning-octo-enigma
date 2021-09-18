package client

import (
	graph "github.com/kiali/kiali/graph/config/cytoscape"
)

func newItem(node *graph.NodeData) Item {
	return Item{Node: node}
}

// MakeGraph returns a Graph variable from a given graphType
func MakeGraph(graphType *graph.Config) Graph {
	adjList := make(Graph)
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

	return adjList
}
