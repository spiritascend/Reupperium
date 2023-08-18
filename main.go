package main

import (
	"net/http"
	"net/url"
	"reupperium/rapidgator"
	"sync"

	"gopkg.in/resty.v1"
)

func main() {
	client := resty.New()

	proxyURL, _ := url.Parse("http://127.0.0.1:8888")
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	client.SetTransport(transport)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		rapidgator.Initialize(client)
	}()

	wg.Wait()
}
