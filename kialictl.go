package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Gituser143/stunning-octo-enigma/pkg/client"
)

func main() {
	// Init a context
	ctx := context.Background()

	// Init Kiali Client
	kc := client.NewKialiClient("localhost", 20001, nil)

	// Get Namspace graph for a namespace
	namespaces := []string{"istio-teastore"}
	graph, err := kc.GetNamespacesGraph(ctx, namespaces)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(graph)
	graph, err = kc.GetWorkloadGraph(ctx, namespaces[0], "teastore-webui")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(graph)

	client.GetMetrics(namespaces[0])

	depName := "teastore-recommender"
	depNamespace := namespaces[0]
	err = client.Scale(depName, depNamespace)

	fmt.Println(err)
}
