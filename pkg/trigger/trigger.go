package trigger

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/Gituser143/stunning-octo-enigma/pkg/k8s"
	client "github.com/Gituser143/stunning-octo-enigma/pkg/kiali-client"
	"github.com/Gituser143/stunning-octo-enigma/pkg/metricscraper"
	"golang.org/x/sync/errgroup"
)

// StartTrigger runs the trigger indefinetely and checks for violations every 30 seconds
func (tc *TriggerClient) StartTrigger(ctx context.Context, thresholds Threshold) error {
	eg, egCtx := errgroup.WithContext(ctx)
	t := time.NewTicker(30 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-t.C:
			eg.Go(func() error {
				return checkThroughput(egCtx, tc.KialiClient, thresholds.Throughput)
			})

			eg.Go(func() error {
				// TODO: get thresholds from above
				return checkResource(egCtx, tc.MetricClient, tc.K8sClient, thresholds.ResourceThresholds)
			})

			if err := eg.Wait(); err != nil {
				if errors.Is(err, errScaleApplication) {
					// Scale App
					log.Println("Scaling Application")
				}
				return err
			}
		}
	}
}

func checkThroughput(ctx context.Context, kc *client.KialiClient, throughput int64) error {
	// Get Namspace graph for a namespace
	namespaces := []string{"istio-teastore"}
	parameters := map[string]string{
		"responseTime": "avg",
		"throughput":   "response",
		"duration":     "1m",
	}

	// Get workload graph for a namespace
	graph, err := kc.GetWorkloadGraph(ctx, namespaces, parameters)
	if err != nil {
		return err
	}

	// Get Unkown ID
	unknownID := ""
	for service, item := range graph {
		if item.Node.Workload == "unknown" {
			unknownID = service
			break
		}
	}

	if unknownID == "" {
		return errors.New("no unkown service")
	}

	item := graph[unknownID]
	currentThroughput := int64(0)

	// Calculate e2e throughput as sum of throughput at unknowns
	for _, edge := range item.Edges {
		t, _ := strconv.ParseInt(edge.Throughput, 10, 64)
		currentThroughput += t
	}

	if currentThroughput < throughput {
		return errScaleApplication
	}

	return nil
}

// TODO
func checkResource(ctx context.Context, mc *metricscraper.MetricClient, kc *k8s.K8sClient, thresholds map[string]Resources) error {
	if thresholds != nil {
		return errScaleApplication
	}
	return nil
}
