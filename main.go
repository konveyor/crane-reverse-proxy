package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
)

type Cluster struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

func main() {

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	client, err := client.New(config, client.Options{})
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)

	ref := types.NamespacedName{
		Namespace: "openshift-migration",
		Name:      "proxy",
	}

	configmap := v1.ConfigMap{}
	err = client.Get(context.TODO(), ref, &configmap)
	if err != nil {
		panic(err)
	}

	clusters := []Cluster{}

	json.Unmarshal([]byte(configmap.Data["clusters"]), &clusters)

	for _, cluster := range clusters {
		var remote *url.URL
		var err error

		ref := types.NamespacedName{
			Namespace: cluster.Namespace,
			Name:      cluster.Name,
		}

		secret := v1.Secret{}
		err = client.Get(context.TODO(), ref, &secret)
		if err != nil {
			panic(err)
		}

		remote, err = url.Parse(string(secret.Data["url"]))
		if err != nil {
			panic(err)
		}

		r.Any("/"+cluster.Namespace+"/"+cluster.Name+"/*proxyPath", func(c *gin.Context) {
			c.Request.URL.Path, _ = c.Params.Get("proxyPath")

			proxy := httputil.NewSingleHostReverseProxy(remote)
			proxy.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			proxy.ServeHTTP(c.Writer, c.Request)
		})

	}

	r.Run(":8080")
}
