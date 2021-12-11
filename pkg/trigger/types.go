package trigger

import "errors"

// TODO
// hardcoding for now, yeet later.
const applicationNamespace = "istio-teastore"

var errScaleApplication = errors.New("scale application")

// Resources holds CPU and Memory values as float64
type Resources struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
}

// Thresholds hold per deployment resource thresholds along with the e2e
// throughput to be maintained for an application
type Thresholds struct {
	ResourceThresholds map[string]Resources `json:"resourceThresholds"`
	Throughput         int64                `json:"throughput"`
}
