package trigger

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
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
					log.Println("trigger client triggered")
					log.Println("fetching base deployments to scale")
					// Scale App
					// Call scale function
					// 	Scale function should calculate replica count of service
					// 	at which trigger occured
					// 	Scale function should then calculate effect of scaling to
					// 	donwstream services
					baseDeps, err := tc.getBaseDeployments(ctx)
					if err != nil && errors.Is(err, errScaleApplication) {
						log.Println("Deployments to scale are:", baseDeps)
						tc.scaleDeployements(ctx, baseDeps)
					}
				} else {
					log.Println("no resource thresholds crossed, not scaling")
					return err
				}
			}
		}
	}
}

func (tc *TriggerClient) getQueueLengthThresholds(fileName string) map[string]float64 {

	queueLengthThresholds := make(map[string]float64)

	jsonFile, err := os.Open(fileName)

	if err != nil {
		log.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal([]byte(byteValue), &queueLengthThresholds)

	return queueLengthThresholds
}

func (tc *TriggerClient) scaleDeployements(ctx context.Context, baseDeps map[string]Resources) error {

	// Initialize an empty list used as a queue for BFS
	graphQueue := list.New()

	// Maintains the replica counts of each service
	oldReplicaCounts := make(map[string]int)
	replicaCounts := make(map[string]int)

	namespaces := []string{applicationNamespace}
	parameters := map[string]string{
		"responseTime": "avg",
		"throughput":   "response",
		"duration":     "1m",
	}

	kialiGraph, err := tc.GetWorkloadGraph(ctx, namespaces, parameters)
	if err != nil {
		return err
	}

	queueLengths, _ := kialiGraph.GetQueueLengths()

	// Initializes the replica count to the current replica count for each service
	for _, item := range kialiGraph {
		if item.Node.Workload == "unknown" {
			continue
		} else {
			currentReplicaCount, err := tc.K8sClient.GetCurrentReplicaCount(ctx, applicationNamespace, item.Node.Workload)

			if err != nil {
				return err
			}

			replicaCounts[item.Node.Workload] = (int)(currentReplicaCount)
			oldReplicaCounts[item.Node.Workload] = (int)(currentReplicaCount)
		}
	}

	baseDependenciesNewReplicaCount, err := tc.getNewReplicaCounts(ctx, baseDeps, applicationNamespace)
	if err != nil {
		return err
	}

	// Iterates through base dependencies
	// Calculates new replica count for them ( based on HPA )
	// Pushes the base dependencies into the graphQueue
	// to perform BFS on it's child nodes
	for service, replicaCount := range baseDependenciesNewReplicaCount {
		log.Printf(
			"[hpa rc for %s] old replica count: %d, new replica count: %d\n",
			service,
			replicaCounts[service],
			replicaCount,
		)
		replicaCounts[service] = (int)(math.Max(float64(replicaCounts[service]), float64(replicaCount)))
		graphQueue.PushBack(service)
	}

	queueLengthThresholds := tc.getQueueLengthThresholds("queue.json")

	idMap := make(map[string]string)
	for id, item := range kialiGraph {
		idMap[item.Node.Workload] = id
	}

	for {
		if graphQueue.Len() == 0 {
			break
		}

		currentNode := graphQueue.Front()
		// curentService here refers to the parent service
		currentService := idMap[fmt.Sprintf("%v", currentNode.Value)]
		graphQueue.Remove(currentNode)

		if _, ok := kialiGraph[currentService]; !ok {
			continue
		}

		for _, edge := range kialiGraph[currentService].Edges {
			if edge == nil {
				continue
			}
			// serviceToScale refers to the child service
			serviceToScale := kialiGraph[edge.Target].Node.Workload
			newQueueLength := queueLengths[serviceToScale] * float64(replicaCounts[currentService]) / float64(oldReplicaCounts[currentService])
			newReplicaCount := (int)(math.Ceil(newQueueLength/queueLengthThresholds[serviceToScale])) * replicaCounts[serviceToScale]

			if newReplicaCount > replicaCounts[serviceToScale] {
				log.Printf(
					"[service: %s] old ql: %f, new ql: %f\n",
					serviceToScale,
					queueLengths[serviceToScale],
					newQueueLength,
				)
				log.Printf(
					"[replicas for: %s] old replica count: %d, new replica count: %d\n\n",
					serviceToScale,
					replicaCounts[serviceToScale],
					newReplicaCount,
				)
				replicaCounts[serviceToScale] = newReplicaCount
				graphQueue.PushBack(serviceToScale)
			}
		}
	}

	for service, replicaCount := range replicaCounts {
		if replicaCount > oldReplicaCounts[service] {
			tc.K8sClient.ScaleDeployment(ctx, applicationNamespace, service, int32(replicaCount))
		}
	}

	return nil
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
// utilization higher than specified threshold along with their respoective metrics
func (tc *TriggerClient) getBaseDeployments(ctx context.Context) (map[string]Resources, error) {
	baseDeps := make(map[string]Resources)

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

	metricsBs, _ := json.MarshalIndent(depMetrics, "", "  ")
	thresholdBs, _ := json.MarshalIndent(tc.thresholds.ResourceThresholds, "", "  ")

	log.Println("Metrics are\n", string(metricsBs))
	log.Println("Thresholds are\n", string(thresholdBs))

	// Check if deployments with thresholds defined have metrics greater than
	// threshold. If yes, that deployment is a base deployment which needs to be
	// scaled.
	for dep, threshold := range tc.thresholds.ResourceThresholds {
		metrics := depMetrics[dep]
		if metrics.CPU > threshold.CPU && threshold.CPU > 0 {
			baseDeps[dep] = metrics
			continue
		}
		if metrics.Memory > threshold.Memory && threshold.Memory > 0 {
			baseDeps[dep] = metrics
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
	if numContainers == 0 {
		return r
	}

	var cpuSum, memSum float64
	for _, c := range metrics.Containers {
		cpuSum += c.Usage.Cpu().AsApproximateFloat64()
		memSum += c.Usage.Memory().AsApproximateFloat64()
	}

	r.CPU = cpuSum / float64(numContainers)
	r.Memory = memSum / float64(numContainers)

	return r
}

// getNewReplicaCounts gets replica counts for problematic deployments,
// i. e., deployments where resource thresholds are exceeded.
// This function returns a map of deployments to their new (HPA) replica counts
func (tc *TriggerClient) getNewReplicaCounts(ctx context.Context, baseDeps map[string]Resources, namespace string) (map[string]int64, error) {

	// desiredReplicas := ceil[currentReplicas * ( currentMetricValue / desiredMetricValue )]

	baseDepsNewReplicaCounts := make(map[string]int64)

	for dep, currentMetrics := range baseDeps {
		currentReplicas, err := tc.K8sClient.GetCurrentReplicaCount(ctx, namespace, dep)
		if err != nil {
			return baseDepsNewReplicaCounts, err
		}

		desiredMetrics := tc.thresholds.ResourceThresholds[dep]

		desiredReplicasCPU := int64(math.Ceil(float64(currentReplicas) * currentMetrics.CPU / desiredMetrics.CPU))
		desiredReplicasMemory := int64(math.Ceil(float64(currentReplicas) * currentMetrics.Memory / currentMetrics.Memory))

		desiredReplicas := int64(0)

		if desiredReplicasCPU > desiredReplicasMemory {
			desiredReplicas = desiredReplicasCPU
		} else {
			desiredReplicas = desiredReplicasMemory
		}

		baseDepsNewReplicaCounts[dep] = desiredReplicas
	}

	return baseDepsNewReplicaCounts, nil
}
