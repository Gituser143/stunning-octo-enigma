package trigger

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// StartTrigger runs the trigger indefinetely and checks for violations every 30 seconds
func (tc *TriggerClient) StartTrigger(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)
	t := time.NewTicker(30 * time.Second)
	thresholds := tc.thresholds

	for {

		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-t.C:
			// Check for throughput violations
			eg.Go(func() error {
				return tc.checkThroughput(egCtx, thresholds.Throughput)
			})

			// Check for resource thresholds exceeding and get corresponding
			// deployments
			eg.Go(func() error {
				return tc.checkResources(egCtx)
			})

			if err := eg.Wait(); err != nil {
				if errors.Is(err, errScaleApplication) {
					// Scale App
					baseDeps, err := tc.getBaseDeployments(ctx)
					if err != nil && !errors.Is(err, errScaleApplication) {
						return err
					}
					log.Println("Deployments to scale are:", baseDeps)
				} else {
					return err
				}
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

func (tc *TriggerClient) checkResources(ctx context.Context) error {
	_, err := tc.getBaseDeployments(ctx)
	return err
}

// getBaseDeployments returns a slice of deployment names that have resource
// utilization higher than specified threshold
func (tc *TriggerClient) getBaseDeployments(ctx context.Context) ([]string, error) {
	baseDeps := []string{}

	pods, err := tc.GetPodNames(ctx, applicationNamespace)
	if err != nil {
		return baseDeps, err
	}

	deps, err := tc.GetDeploymentNames(ctx, applicationNamespace)
	if err != nil {
		return baseDeps, err
	}

	// create a mapping of deployments -> pods belonging to that
	// deployment.
	depsToPods := make(map[string][]string)
	for _, dep := range deps {
		depsToPods[dep] = getPodsForDeployment(dep, pods)
	}

	// get current deployment metrics
	depMetrics := tc.getPerDeploymentMetrics(ctx, depsToPods)

	// Check if deployments with thresholds defined have metrics greater than
	// threshold. If yes, that deployment is a base deployment which needs to be
	// scaled.
	for dep, threshold := range tc.thresholds.ResourceThresholds {
		metrics := depMetrics[dep]
		if metrics.CPU > threshold.CPU && threshold.CPU > 0 {
			baseDeps = append(baseDeps, dep)
			continue
		}
		if metrics.Memory > threshold.Memory && threshold.Memory > 0 {
			baseDeps = append(baseDeps, dep)
			continue
		}
	}

	if len(baseDeps) > 0 {
		return baseDeps, errScaleApplication
	}

	return baseDeps, nil
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
func getPodsForDeployment(deploymentName string, pods []string) []string {
	resPods := []string{}

	for _, pod := range pods {
		if strings.HasPrefix(pod, deploymentName+"-") {
			resPods = append(resPods, pod)
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
