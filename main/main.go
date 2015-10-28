package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	rancher "github.com/rancher/go-rancher/client"
)

func main() {
	// Simple example of using this library. Run this as follows
	//
	//     go run main/main.go myhost /tmp/myhost.sock
	//
	// Then run `docker -H unix:///tmp/myhost.sock ps`

	if err := run(); err != nil {
		logrus.Fatal(err)
	}
}

func run() error {
	logrus.SetLevel(logrus.DebugLevel)

	client, err := rancher.NewRancherClient(&rancher.ClientOpts{
		Url:       os.Getenv("CATTLE_URL"),
		AccessKey: os.Getenv("CATTLE_ACCESS_KEY"),
		SecretKey: os.Getenv("CATTLE_SECRET_KEY"),
	})
	if err != nil {
		return err
	}

	proxy := dockerapiproxy.NewProxy(client, os.Args(1), os.Args(2))
	return proxy.ListenAndServe()
}
