package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
)

const (
	proxySecretName = "crane-proxy"
)

type Cluster struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

func main() {

	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		log.Fatalf("Please set the 'NAMESPACE' environment variable.")
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Unable to retrieve in cluster kubeconfig.")
	}

	client, err := client.New(config, client.Options{})
	if err != nil {
		log.Fatalf("Unable to create kubernetes client.")
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)

	ref := types.NamespacedName{
		Namespace: namespace,
		Name:      proxySecretName,
	}

	configmap := v1.ConfigMap{}
	err = client.Get(context.TODO(), ref, &configmap)
	if err != nil {
		log.Fatalf("Unable to load ConfigMap: %s in Namespace: %s", proxySecretName, namespace)
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
			log.Printf("Unable to retrieve secret %s in Namespace: %s. No proxy created for this clsuter.", cluster.Name, cluster.Namespace)
		}

		remote, err = url.Parse(string(secret.Data["url"]))
		if err != nil {
			log.Printf("No URL found in secret %s in Namespace: %s. No proxy created for this clsuter.", cluster.Name, cluster.Namespace)
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
