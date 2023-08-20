package main

import (
	"fmt"
	"net/http"
	"net/url"
	"reupperium/filecrypt"
	"reupperium/utils"

	"gopkg.in/resty.v1"
)

func main() {

	_, err := utils.CopyAll()

	if err != nil {
		fmt.Println(err)
	}

	client := resty.New()

	proxyURL, _ := url.Parse("http://127.0.0.1:8888")
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	client.SetTransport(transport)

	filecrypt.Initialize(client)
}
