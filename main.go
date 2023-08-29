package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reupperium/utils"
	"time"

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
	for {
		fmt.Println("Running Update Check")
		err := UpdateCheckV2(restyclient, httpclient)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(1 * time.Hour)
	}
}
