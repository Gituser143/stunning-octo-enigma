package main

import (
	"context"
	"log"

	"github.com/Gituser143/stunning-octo-enigma/pkg/k8s"
	client "github.com/Gituser143/stunning-octo-enigma/pkg/kiali-client"
	"github.com/Gituser143/stunning-octo-enigma/pkg/metricscraper"
	trigger "github.com/Gituser143/stunning-octo-enigma/pkg/trigger"
)

func main() {
	// Init a context
	ctx := context.Background()

	// Variables for Kiali Client
	host := "localhost"
	port := 8081

	// Init Kiali Client
	kc := client.NewKialiClient(host, port, nil)

	// Init Metrics Client
	mc, err := metricscraper.NewMetricClient()
	if err != nil {
		log.Fatal("failied to init metrics client")
	}

	// Init K8s Client
	k8sc, err := k8s.NewK8sClient()
	if err != nil {
		log.Fatal("failied to init k8s client")
	}

	// Init Trigger Client
	tc := trigger.TriggerClient{
		KialiClient:  kc,
		MetricClient: mc,
		K8sClient:    k8sc,
	}

	// TODO: Fetch thresholds

	// Run Trigger
	err = tc.StartTrigger(ctx, trigger.Threshold{})
	if err != nil {
		log.Fatal(err)
	}
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
