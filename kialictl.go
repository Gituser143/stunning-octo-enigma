package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	client "github.com/Gituser143/stunning-octo-enigma/pkg/kiali-client"
)

func main() {
	// Init a context
	ctx := context.Background()

	// Init Kiali Client
	kc := client.NewKialiClient("localhost", 8081, nil)

	// Get Namspace graph for a namespace
	namespaces := []string{"istio-teastore"}
	parameters := map[string]string{
		"duration":     "3h",
		"responseTime": "avg",
	}

	// Get workload graph for a namespace
	graph, err := kc.GetWorkloadGraph(ctx, namespaces, parameters)
	if err != nil {
		log.Println(err)
	} else {
		bs, err := json.Marshal(graph)
		if err != nil {
			log.Println(err)
		} else {
			fmt.Println(string(bs))
		}
	}

	// Get workload metrics for a workload and namespace
	metrics, err := kc.GetWorkloadMetrics(ctx, namespaces[0], "teastore-webui")
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(metrics)
	}
}
