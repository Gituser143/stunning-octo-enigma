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

	"github.com/Gituser143/stunning-octo-enigma/pkg/k8s"
	client "github.com/Gituser143/stunning-octo-enigma/pkg/kiali-client"
	"github.com/Gituser143/stunning-octo-enigma/pkg/metricscraper"
	"github.com/Gituser143/stunning-octo-enigma/pkg/trigger"

	flag "github.com/spf13/pflag"
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

	// Fetch thresholds

	var threshold trigger.Threshold

	filePath := flag.String("file", ".", "Path to config file or directory")
	flag.Parse()

	fileInfo, err := os.Stat(*filePath)

	if err != nil {
		log.Fatal(err)
	}

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
					err = json.Unmarshal(bs, &threshold)
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
		err = json.Unmarshal(bs, &threshold)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println(threshold)

	// Run Trigger
	err = tc.StartTrigger(ctx, threshold)
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
