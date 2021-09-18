package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	client "github.com/Gituser143/stunning-octo-enigma/pkg/kiali-client"
	load "github.com/Gituser143/stunning-octo-enigma/pkg/load-generator"
)

func main() {
	// Init a context
	ctx := context.Background()

	// Init Kiali Client
	kc := client.NewKialiClient("localhost", 20001, nil)

	// Get Namspace graph for a namespace
	namespaces := []string{"istio-teastore"}
	parameters := map[string]string{
		"responseTime": "avg",
		"throughput":   "response",
	}

	exit := make(chan int)

	go func() {
		t := time.NewTicker(3 * time.Second)
		for {
			select {
			case <-t.C:
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
			case <-exit:
				return
			}
		}
	}()

	// Get workload metrics for a workload and namespace
	// metrics, err := kc.GetWorkloadMetrics(ctx, namespaces[0], "teastore-webui")
	// if err != nil {
	// 	log.Println(err)
	// } else {
	// 	fmt.Println(metrics)
	// }

	// Init Load Generator client
	sc := load.NewStressClient("http", "localhost", 30080, nil)
	sc.SetTargetFunction(sc.GetTeaStoreTargets)

	// Begin stress test
	sc.StressApplication("inc", 100, 1, 5, 10, 100)
	time.Sleep(30 * time.Second)
	exit <- 0
}

/*
// Example usage of MetricClient

func exampleUsageMetricClient() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	client, err := ms.NewMetricClient()
	if err != nil {
		return err
	}

	podMetrics := client.StreamPodMetrics(ctx, "default", "teastore-auth-7947675f98-knvq5", 1*time.Second)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case metric := <-podMetrics:
			fmt.Println(metric.Containers[0].Usage.Cpu())
		}
	}
}
*/
