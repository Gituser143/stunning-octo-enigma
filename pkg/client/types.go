package client

import (
	"net/http"
)

type KialiClient struct {
	httpClient *http.Client
	host       string
	port       int
}

func NewKialiClient(host string, port int, hc *http.Client) KialiClient {
	kc := KialiClient{
		host: host,
		port: port,
	}
	if hc != nil {
		kc.httpClient = hc
	} else {
		kc.httpClient = http.DefaultClient
	}

	return kc
}
