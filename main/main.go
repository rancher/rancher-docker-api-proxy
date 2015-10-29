package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
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

func loadTLSConfig(ca, cert, key string) (*tls.Config, error) {
	c, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	config := &tls.Config{
		RootCAs:      pool,
		Certificates: []tls.Certificate{c},
		MinVersion:   tls.VersionTLS10,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
	}

	f, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, err
	}
	config.RootCAs.AppendCertsFromPEM(f)

	return config, nil
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

	tlsConfig, _ := loadTLSConfig("ca.pem", "server-cert.pem", "server-key.pem")

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
			proxy.TlsConfig = tlsConfig
			if err := proxy.ListenAndServe(); err != nil {
				logrus.Errorf("Error on %s [%s]: %v", host, listen, err)
			}
		}(i)
	}

	wg.Wait()
	return nil
}
