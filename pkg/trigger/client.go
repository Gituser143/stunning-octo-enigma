package trigger

import (
	"github.com/Gituser143/stunning-octo-enigma/pkg/k8s"
	"github.com/Gituser143/stunning-octo-enigma/pkg/kiali"
	"github.com/Gituser143/stunning-octo-enigma/pkg/metricscraper"
)

// Client encapsulates a kiali client, metrics server client and a k8s
// client. It is used to trigger scaling for an application
type Client struct {
	KialiClient  *kiali.Client
	MetricClient *metricscraper.Client
	K8sClient    *k8s.Client
	thresholds   Thresholds
}

// SetThresholds sets the thresholds for a given trigger client
func (tc *Client) SetThresholds(thresholds Thresholds) {
	for k, v := range thresholds.ResourceThresholds {
		thresholds.ResourceThresholds[k] = Resources{CPU: v.CPU / 1000}
	}
	tc.thresholds = thresholds
}
