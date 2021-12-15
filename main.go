package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jmontleon/reverse-proxy-poc/pkg/rproxy"
)

func main() {
	proxy, err := rproxy.NewProxy(os.Getenv("CLUSTER_URL"))
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", rproxy.ProxyRequestHandler(proxy))
	log.Fatal(http.ListenAndServeTLS(":8080",
		os.Getenv("TLSCrt"),
		os.Getenv("TLSKey"),
		nil))
}
