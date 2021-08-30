package main

import (
	"context"
	"log"

	"github.com/Gituser143/stunning-octo-enigma/pkg/client"
)

func main() {
	// Init a context
	ctx := context.Background()

	// Init Kiali Client
	kc := client.NewKialiClient(nil)

	// Get Namspace graph for a namespace
	namespaces := []string{"istio-teastore"}
	err := kc.GetNamespacesGraph(ctx, "localhost", 20001, namespaces)
	if err != nil {
		log.Fatal(err)
	}
}
