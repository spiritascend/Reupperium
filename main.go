package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"reupperium/utils"
	"runtime"
	"time"

	_ "net/http/pprof"

	"gopkg.in/resty.v1"
)

func main() {
	config, err := utils.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		runtime.SetBlockProfileRate(1)
		runtime.SetMutexProfileFraction(10)
		_ = http.ListenAndServe("localhost:6060", nil)
	}()
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
		go func() {
			updatecheckstart := time.Now()
			err := UpdateCheckV2(restyclient, httpclient)
			if err != nil {
				fmt.Println(err)
			}
			timesinceupdatecheckstart := time.Since(updatecheckstart)
			fmt.Printf("Update Check took %s\n", timesinceupdatecheckstart)
		}()
		time.Sleep(time.Duration(config.TimeBeforeNextDeletedCheckMs) * time.Millisecond)
	}

}
