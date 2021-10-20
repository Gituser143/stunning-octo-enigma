package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Gituser143/stunning-octo-enigma/pkg/config"
	"github.com/Gituser143/stunning-octo-enigma/pkg/k8s"
	"github.com/Gituser143/stunning-octo-enigma/pkg/kiali"
	load "github.com/Gituser143/stunning-octo-enigma/pkg/load-generator"
	"github.com/Gituser143/stunning-octo-enigma/pkg/metricscraper"
	"github.com/Gituser143/stunning-octo-enigma/pkg/trigger"

	flag "github.com/spf13/pflag"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {

	// Flag definitions
	loadtest := flag.BoolP("load", "l", false, "Load test application")
	filePath := flag.StringP("file", "f", "config.json", "Path to config file or directory")
	scaleAndLoad := flag.BoolP("scale-and-load", "s", false, "Running scaler and simultaneously load test application")
	shouldLogReplicaCounts := flag.BoolP("logrc", "r", false, "Log replica counts of application deployments to file (use alongside l or s)")
	shouldLogThroughput := flag.BoolP("logth", "t", false, "Log e2e throughput of application (use alongside l or s)")
	shouldLogQueueLens := flag.BoolP("logq", "q", false, "Log queue lengths and create json with threshold queue lengths for each deployment of application (use alongside l)")
	flag.Parse()

	// Get config from config file
	conf, err := config.GetConfig(*filePath)
	if err != nil {
		if errors.Is(err, config.ErrInavlidConfigPath) {
			flag.Usage()
		}
		log.Fatal(err)
	}
	log.Println("read config file")

	// Init a context
	ctx := context.Background()

	// Init Kiali Client
	kc := kiali.NewKialiClient(conf.KialiHost.Host, conf.KialiHost.Port, nil)

	// Init Metrics Client
	mc, err := metricscraper.NewMetricClient()
	if err != nil {
		log.Fatal("failed to init metrics client: ", err)
	}

	// Init K8s Client
	k8sc, err := k8s.NewK8sClient()
	if err != nil {
		log.Fatal("failed to init k8s client")
	}

	// Init Trigger Client
	tc := trigger.Client{
		KialiClient:  kc,
		MetricClient: mc,
		K8sClient:    k8sc,
	}
	tc.SetThresholds(conf.Thresholds)
	log.Println("initialised trigger client")

	if *loadtest {
		loadTest(ctx, &tc, conf, *shouldLogQueueLens, *shouldLogReplicaCounts, *shouldLogThroughput)
	} else if *scaleAndLoad {
		// Run load test
		go loadTest(ctx, &tc, conf, *shouldLogQueueLens, *shouldLogReplicaCounts, *shouldLogThroughput)

		// Run Trigger
		err = tc.StartTrigger(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func printThrpughput(ctx context.Context, tc *trigger.Client) {
	f, err := os.OpenFile("throughput_load.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// // log.Println(err)
	}

	defer f.Close()

	logTicker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return

		case <-logTicker.C:
			throughput, err := tc.GetE2EThroughput(ctx)
			if err != nil {
				// log.Println("error getting e2e throughput:", err)
			} else {
				ts := fmt.Sprintf("%d,%v\n", throughput, time.Now())
				if _, err := f.WriteString(ts); err != nil {
					// // log.Println(err)
				}
			}
		}
	}
}

func printReplicaCount(ctx context.Context, tc *trigger.Client, namespace string) {
	f, err := os.OpenFile("replica_counts.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// log.Println(err)
	}

	defer f.Close()

	deployments, _ := tc.K8sClient.GetDeploymentNames(ctx, namespace)
	logTicker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return

		case <-logTicker.C:
			for _, dep := range deployments {
				replicaCount, _ := tc.K8sClient.GetCurrentReplicaCount(ctx, namespace, dep)
				ts := fmt.Sprintf("%s,%d,%v\n", dep, replicaCount, time.Now())
				if _, err := f.WriteString(ts); err != nil {
					// log.Println(err)
				}
			}
		}
	}
}

func loadTest(
	ctx context.Context,
	tc *trigger.Client,
	conf config.Config,
	shouldLogQueueLens bool,
	shouldLogReplicaCounts bool,
	shouldLogThroughput bool) {
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()

	if conf.AppHost.Scheme == "" {
		conf.AppHost.Scheme = "http"
	}

	// Init stress client
	sc := load.NewStressClient(conf.AppHost.Scheme, conf.AppHost.Host, conf.AppHost.Port, nil)
	sc.SetTargetFunction(sc.GetTeaStoreTargets)

	parameters := map[string]string{
		"responseTime": "avg",
		"throughput":   "response",
		"duration":     "1m",
	}

	exitChan := make(chan int)

	if shouldLogQueueLens {
		// Write queue lengths to file
		go logQueuelengths(ctx, tc, conf.Namespaces, parameters, exitChan, &wg)
	}

	if shouldLogReplicaCounts {
		// Print replica counts every 5 seconds to file
		go printReplicaCount(ctx, tc, conf.Namespaces[0])
	}

	if shouldLogThroughput {
		go printThrpughput(ctx, tc)
	}

	// Begin stress test
	sc.StressApplication(conf.LoadConfig)

	log.Println("Finished load testing on application")
	time.Sleep(10 * time.Second)
	exitChan <- 1
}

func logQueuelengths(
	ctx context.Context,
	tc *trigger.Client,
	namespaces []string,
	parameters map[string]string,
	exitChan chan int,
	wg *sync.WaitGroup,
) {
	maxQueueLengths := make(map[string]float64)

	f, err := os.OpenFile("queue_lengths.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// // log.Println(err)
	}

	defer f.Close()

	// Dump Max Queue Lengths to a file
	defer func() {
		defer wg.Done()
		bs, err := json.MarshalIndent(maxQueueLengths, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile("queue.json", bs, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Create ticker to get metrics every 3 seconds
	t := time.NewTicker(3 * time.Second)

	for {
		egCtx, cancel := context.WithCancel(ctx)
		select {
		case <-t.C:
			// Get workload graph for a namespace
			graph, err := tc.KialiClient.GetWorkloadGraph(egCtx, namespaces, parameters)
			if err != nil {
				// // log.Println(err)
			} else {
				// Extract queue lengths from graph
				queueLengths, _ := graph.GetQueueLengths()

				// Store max queue lengths only
				for dep, q := range queueLengths {
					if maxQ, ok := maxQueueLengths[dep]; !ok {
						maxQueueLengths[dep] = q
					} else {
						if q > maxQ {
							maxQueueLengths[dep] = q
						}
					}
					qs := fmt.Sprintf("%s,%f,%v\n", dep, q, time.Now())
					if _, err := f.WriteString(qs); err != nil {
						// // log.Println(err)
					}
				}
			}

		case <-exitChan:
			cancel()
			return

			// case err := <-egCtx.Done():
			// log.Println("Context cancelled", err)
		}
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
