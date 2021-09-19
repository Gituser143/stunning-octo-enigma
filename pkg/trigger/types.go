package trigger

import "errors"

// hardcoding for now, yeet later.
const applicationNamespace = "istio-teastore"

var errScaleApplication = errors.New("scale application")

type Resources struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
}

type Threshold struct {
	ResourceThresholds map[string]Resources `json:"thresholds"`
	Throughput         int64                `json:"throughput"`
}
