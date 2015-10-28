package main

import (
	"os"
	"sync"

	"github.com/Sirupsen/logrus"
	rancher "github.com/rancher/go-rancher/client"
	"github.com/rancher/rancher-docker-api-proxy"
)

func main() {
	// Simple example of using this library. Run this as follows
	//
	//     go run main/main.go myhost unix:///tmp/myhost.sock
	//
	// Then run `docker -H unix:///tmp/myhost.sock ps`

	if err := run(); err != nil {
		logrus.Fatal(err)
	}
}

func run() error {
	if os.Getenv("PROXY_DEBUG") != "" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	client, err := rancher.NewRancherClient(&rancher.ClientOpts{
		Url:       os.Getenv("CATTLE_URL"),
		AccessKey: os.Getenv("CATTLE_ACCESS_KEY"),
		SecretKey: os.Getenv("CATTLE_SECRET_KEY"),
	})
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	for i := range os.Args {
		if len(os.Args) <= i+1 || i%2 == 0 {
			continue
		}

		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			host := os.Args[i]
			listen := os.Args[i+1]
			logrus.Infof("Proxying %s on %s", host, listen)
			proxy := dockerapiproxy.NewProxy(client, host, listen)
			if err := proxy.ListenAndServe(); err != nil {
				logrus.Errorf("Error on %s [%s]: %v", host, listen, err)
			}
		}(i)
	}

	wg.Wait()
	return nil
}
