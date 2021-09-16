package load

import (
	"fmt"
	"net/http"

	vegeta "github.com/tsenart/vegeta/lib"
)

// StressClient is a type through which load testing for a given application can be performed
type StressClient struct {
	scheme     string
	httpClient *http.Client
	host       string
	getTargets func() []vegeta.Target
}

// NewStressClient is a constructor for StressClient
func NewStressClient(scheme, host string, port int, hc *http.Client) StressClient {
	lc := StressClient{
		host:   fmt.Sprintf("%s:%d", host, port),
		scheme: scheme,
	}

	if hc != nil {
		lc.httpClient = hc
	} else {
		lc.httpClient = http.DefaultClient
	}
	return lc
}

// SetTargetFunction is used to set the getTargets function for a StressClient.
// The function should generate the endpoints to b e targetted as a list of vegeta targets
func (lc *StressClient) SetTargetFunction(targetFunc func() []vegeta.Target) {
	lc.getTargets = targetFunc
}
