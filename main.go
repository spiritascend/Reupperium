package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reupperium/utils"

	"gopkg.in/resty.v1"
)

func main() {
	restyclient := resty.New()
	httpclient := &http.Client{}
	for _, arg := range os.Args {
		if arg == "-proxy" {
			winprox, err := utils.GetWindowsProxy()
			if err != nil {
				fmt.Println(err)
				return
			}
			proxyURL, _ := url.Parse("http://" + winprox)
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
			restyclient.SetTransport(transport)
			httpclient.Transport = transport
			break
		}
	}

	err := UpdateCheck(restyclient, httpclient)

	if err != nil {
		fmt.Println(err)
	}
}
