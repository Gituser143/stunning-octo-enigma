package trigger

import "errors"

// hardcoding for now, yeet later.
const applicationNamespace = "istio-teastore"

var errScaleApplication = errors.New("scale application")

type Resources struct {
	CPU    float64
	Memory float64
}

type Threshold struct {
	ResourceThresholds map[string]Resources
	Throughput         int64
}
