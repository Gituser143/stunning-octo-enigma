package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Gituser143/stunning-octo-enigma/pkg/k8s"
	client "github.com/Gituser143/stunning-octo-enigma/pkg/kiali-client"
	load "github.com/Gituser143/stunning-octo-enigma/pkg/load-generator"
	"github.com/Gituser143/stunning-octo-enigma/pkg/metricscraper"
	"github.com/Gituser143/stunning-octo-enigma/pkg/trigger"

	flag "github.com/spf13/pflag"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	// Init a context
	ctx := context.Background()

	// Variables for Kiali Client
	host := "localhost"
	port := 20001

	// Init Kiali Client
	kc := client.NewKialiClient(host, port, nil)

	// Init Metrics Client
	mc, err := metricscraper.NewMetricClient()
	if err != nil {
		log.Fatal("failied to init metrics client: ", err)
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

	// Check if load test phase
	loadtest := flag.BoolP("load", "l", false, "Load test or not")

	// Get metrics config file path
	filePath := flag.StringP("file", "f", "", "Path to config file or directory")

	scaleAndLoad := flag.BoolP("scale-and-load", "s", false, "running scaler + simulataneously load test shit")

	flag.Parse()

	if *loadtest {
		loadTest(ctx, &tc, true)
	} else {
		if *filePath == "" {
			flag.Usage()
			os.Exit(1)
		}

		fileInfo, err := os.Stat(*filePath)
		if err != nil {
			log.Fatal(err)
		}

		// Fetch thresholds
		var thresholds trigger.Thresholds

		// TODO: Handle multiple file configs cleanly
		if fileInfo.IsDir() {
			err = filepath.Walk(*filePath,
				func(path string, info fs.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() {
						bs, err := ioutil.ReadFile(path)
						if err != nil {
							fmt.Println("2")
							return err
						}
						err = json.Unmarshal(bs, &thresholds)
						if err != nil {
							return err
						}
					}
					return nil
				})
			if err != nil {
				log.Fatal(err)
			}
		} else {
			bs, err := ioutil.ReadFile(*filePath)
			if err != nil {
				log.Fatal(err)
			}
			err = json.Unmarshal(bs, &thresholds)
			if err != nil {
				log.Fatal(err)
			}
		}

		log.Println("read config file")

		tc.SetThresholds(thresholds)

		log.Println("initialised trigger client")

		if *scaleAndLoad {
			go loadTest(ctx, &tc, false)
		}
		// Run Trigger
		err = tc.StartTrigger(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func loadTest(ctx context.Context, tc *trigger.TriggerClient, shouldLogQueueLens bool) {
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()

	// TODO: Get load test parameters from configs
	scheme := "http"
	host := os.Getenv("TEASTORE_HOST")
	if host == "" {
		log.Fatal("TEASTORE_HOST env not set")
	}
	port := 8080

	distributionType := "inc"
	steps := 50
	duration := 15
	workers := 50
	minRate := 50
	maxRate := 500

	namespaces := []string{"istio-teastore"}

	// Init stress client
	sc := load.NewStressClient(scheme, host, port, nil)
	sc.SetTargetFunction(sc.GetTeaStoreTargets)

	parameters := map[string]string{
		"responseTime": "avg",
		"throughput":   "response",
		"duration":     "1m",
	}

	exitChan := make(chan int)

	if shouldLogQueueLens {
		// Write queue lengths to file
		go logQueuelengths(ctx, tc, namespaces, parameters, exitChan, &wg)
	}

	// Begin stress test
	sc.StressApplication(distributionType, steps, duration, workers, minRate, maxRate)

	time.Sleep(10 * time.Second)
	exitChan <- 1
}

func logQueuelengths(
	ctx context.Context,
	tc *trigger.TriggerClient,
	namespaces []string,
	parameters map[string]string,
	exitChan chan int,
	wg *sync.WaitGroup,
) {
	maxQueueLengths := make(map[string]float64)

	// Dump Queue Lengths to a file
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
		select {
		case <-t.C:
			// Get workload graph for a namespace
			graph, err := tc.GetWorkloadGraph(ctx, namespaces, parameters)
			if err != nil {
				log.Println(err)
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
				}
			}

		case <-exitChan:
			return

		case <-ctx.Done():
			return
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
