package client

import (
	"fmt"
	"net/http"
)

// KialiClient is a type to help interact with kiali dashboards
type KialiClient struct {
	httpClient *http.Client
	host       string
}

// NewKialiClient is a constructor for type KialiClient
func NewKialiClient(host string, port int, hc *http.Client) KialiClient {
	kc := KialiClient{
		host: fmt.Sprintf("%s:%d", host, port),
	}

	if hc != nil {
		kc.httpClient = hc
	} else {
		kc.httpClient = http.DefaultClient
	}
	return kc
}
