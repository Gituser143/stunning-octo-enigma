package client

import (
	"fmt"
	"net/http"
)

type KialiClient struct {
	httpClient *http.Client
	host       string
}

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
