package trigger

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
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
				return tc.checkThroughput(egCtx, thresholds.Throughput)
			})

			eg.Go(func() error {
				// TODO: get thresholds from above
				return tc.checkResource(egCtx, thresholds.ResourceThresholds)
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

func (tc *TriggerClient) checkThroughput(ctx context.Context, throughput int64) error {
	// Get Namspace graph for a namespace
	namespaces := []string{applicationNamespace}
	parameters := map[string]string{
		"responseTime": "avg",
		"throughput":   "response",
		"duration":     "1m",
	}

	// Get workload graph for a namespace
	graph, err := tc.GetWorkloadGraph(ctx, namespaces, parameters)
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

func (tc *TriggerClient) checkResource(ctx context.Context, thresholds map[string]Resources) error {
	pods, err := tc.GetPods(ctx, applicationNamespace)
	if err != nil {
		return err
	}

	deps, err := tc.GetDeploymentNames(ctx, applicationNamespace)
	if err != nil {
		return err
	}

	// create a mapping of deployments -> pods belonging to that
	// deployment.
	depsToPods := make(map[string][]string)
	for _, dep := range deps {
		depsToPods[dep] = getPodsForDeployment(dep, pods)
	}

	metricsPerDeployment := tc.getPerDeploymentMetrics(ctx, depsToPods)
	if decideIfResourceTriggerShouldHappen(metricsPerDeployment, thresholds) {
		return errScaleApplication
	}

	return nil
}

func (tc *TriggerClient) getPerDeploymentMetrics(ctx context.Context, depPodMap map[string][]string) map[string]Resources {
	resourceMap := make(map[string]Resources)
	for dep, pods := range depPodMap {
		for _, pod := range pods {
			metrics, err := tc.GetPodMetrics(ctx, applicationNamespace, pod)
			if err != nil {
				// print out error and not return to try and get as many pod metrics
				// as possible.
				log.Printf("error %s getting metrics of pod %s", err.Error(), pod)
			}
			resourceMap[dep] = aggregatePodMetricsToResources(metrics)
		}
	}

	return resourceMap
}

// this is a super hacky way of doing this, but we shall make our peace with it for now.
func getPodsForDeployment(deploymentName string, pods []apiv1.Pod) []string {
	resPods := []string{}

	for _, pod := range pods {
		if strings.HasPrefix(pod.Name, deploymentName+"-") {
			resPods = append(resPods, pod.Name)
		}
	}

	return resPods
}

func aggregatePodMetricsToResources(metrics *v1beta1.PodMetrics) Resources {
	r := Resources{}
	numContainers := len(metrics.Containers)

	var cpuSum, memSum float64
	for _, c := range metrics.Containers {
		cpuSum += c.Usage.Cpu().AsApproximateFloat64()
		memSum += c.Usage.Memory().AsApproximateFloat64()
	}

	r.CPU = cpuSum / float64(numContainers)
	r.Memory = memSum / float64(numContainers)

	return r
}

// TODO: we need to discuss the reactive trigger and then implement this - for now, it'll decide to trigger
// even if one deployment's CPU usage crosses threshold.
func decideIfResourceTriggerShouldHappen(metricsPerDeployment map[string]Resources, thresholds map[string]Resources) bool {
	for dep, metrics := range metricsPerDeployment {
		thresh := thresholds[dep]
		if thresh.CPU <= metrics.CPU {
			return true
		}
	}

	return false
}
