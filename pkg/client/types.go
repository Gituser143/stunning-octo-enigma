package client

import (
	"net/http"
)

type KialiClient struct {
	HttpClient *http.Client
}

func NewKialiClient(hc *http.Client) KialiClient {
	kc := KialiClient{}
	if hc != nil {
		kc.HttpClient = hc
	} else {
		kc.HttpClient = http.DefaultClient
	}

	return kc
}
