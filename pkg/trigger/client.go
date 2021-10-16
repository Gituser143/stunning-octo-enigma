package trigger

import (
	"github.com/Gituser143/stunning-octo-enigma/pkg/k8s"
	client "github.com/Gituser143/stunning-octo-enigma/pkg/kiali-client"
	"github.com/Gituser143/stunning-octo-enigma/pkg/metricscraper"
)

type TriggerClient struct {
	*client.KialiClient
	*metricscraper.MetricClient
	*k8s.K8sClient
	thresholds Thresholds
}

// SetThresholds sets the thresholds for a given trigger client
func (tc *TriggerClient) SetThresholds(thresholds Thresholds) {
	for k, v := range thresholds.ResourceThresholds {
		thresholds.ResourceThresholds[k] = Resources{CPU: v.CPU / 1000}
	}
	tc.thresholds = thresholds
}
