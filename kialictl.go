package main

import (
	"context"
	"log"

	ms "github.com/Gituser143/stunning-octo-enigma/pkg/metricscraper"
)

func main() {
	// Init a context
	ctx := context.Background()

	// // Init Kiali Client
	// kc := client.NewKialiClient("localhost", 20001, nil)

	// // Get Namspace graph for a namespace
	namespaces := []string{"istio-teastore"}
	// graph, err := kc.GetNamespacesGraph(ctx, namespaces)
	// if err != nil {
	// 	log.Println(err)
	// } else {
	// 	fmt.Println(graph)
	// }

	// graph, err = kc.GetWorkloadGraph(ctx, namespaces[0], "teastore-webui")
	// if err != nil {
	// 	log.Println(err)
	// } else {
	// 	fmt.Println(graph)
	// }

	// depName := "teastore-recommender"
	// depNamespace := namespaces[0]
	// 	err := client.Scale(depName, depNamespace)

	// 	fmt.Println(err)

	err := ms.LogContainerMetrics(ctx, namespaces[0], 500)
	if err != nil {
		log.Fatal(err)
	}
}
