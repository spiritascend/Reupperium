package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"reupperium/utils"
	"runtime"
	"sync"
	"time"

	_ "net/http/pprof"
)

func main() {
	config, err := utils.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	httpclient := &http.Client{Timeout: time.Duration(config.MaxTimeoutUploadBeforeRetryMs) * time.Millisecond}
	for _, arg := range os.Args {

		if arg == "-cpuprofile" {
			go func() {
				runtime.SetBlockProfileRate(1)
				runtime.SetMutexProfileFraction(10)
				_ = http.ListenAndServe("localhost:"+config.CPUProfilePort, nil)
			}()
			fmt.Printf("cpuprofile listening on port %s go to http://localhost:%s/debug/pprof/ to view the profile\n", config.CPUProfilePort, config.CPUProfilePort)
		}

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
			httpclient.Transport = transport
			break
		}

	}

	for {
		var mfnwg sync.WaitGroup
		mfnwg.Add(1)

		go func() {
			defer mfnwg.Done()
			updatecheckstart := time.Now()
			err := UpdateCheckV2(httpclient)
			if err != nil {
				fmt.Println(err)
			}
			timesinceupdatecheckstart := time.Since(updatecheckstart)
			fmt.Printf("Update Check took %s\n", timesinceupdatecheckstart)
		}()

		mfnwg.Wait()

		time.Sleep(time.Duration(config.TimeBeforeNextDeletedCheckMs) * time.Millisecond)
	}

}
