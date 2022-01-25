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
	clusterMap := make(map[string]*url.URL)

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Unable to retrieve in cluster kubeconfig.")
	}

	client, err := client.New(config, client.Options{})
	if err != nil {
		log.Fatalf("Unable to create kubernetes client.")
	}

	clusterMapInit(client, &clusterMap)

	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.Any("/:namespace/:name/*proxyPath", func(c *gin.Context) {
		var proxy *httputil.ReverseProxy

		namespace, _ := c.Params.Get("namespace")
		name, _ := c.Params.Get("name")

		if url, ok := clusterMap[namespace+name]; ok {
			proxy = httputil.NewSingleHostReverseProxy(url)
		} else {
			url := clusterMapAdd(client, &clusterMap, namespace, name)
			if url != nil {
				proxy = httputil.NewSingleHostReverseProxy(url)
			}
		}

		if proxy == nil {
			c.AbortWithStatus(http.StatusBadGateway)
		} else {
			proxy.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}

			c.Request.URL.Path, _ = c.Params.Get("proxyPath")
			c.Request.Host = clusterMap[namespace+name].Host

			proxy.ServeHTTP(c.Writer, c.Request)

			if c.Writer.Status() == http.StatusBadGateway {
				clusterMapRemove(&clusterMap, namespace, name)
			}
		}
	})

	r.Run(":8080")
}

func clusterMapInit(client client.Client, clusterMap *map[string]*url.URL) {

	namespace := os.Getenv("NAMESPACE")
	configmap := v1.ConfigMap{}
	clusters := []Cluster{}

	if namespace == "" {
		log.Fatalf("Please set the 'NAMESPACE' environment variable.")
	}

	ref := types.NamespacedName{
		Namespace: namespace,
		Name:      proxySecretName,
	}

	err := client.Get(context.TODO(), ref, &configmap)
	if err != nil {
		log.Fatalf("Unable to load ConfigMap: %s in Namespace: %s", proxySecretName, namespace)
	}

	json.Unmarshal([]byte(configmap.Data["clusters"]), &clusters)

	for _, cluster := range clusters {
		ref := types.NamespacedName{
			Namespace: cluster.Namespace,
			Name:      cluster.Name,
		}

		secret := v1.Secret{}
		err := client.Get(context.TODO(), ref, &secret)

		if err != nil {
			log.Printf("Unable to retrieve secret %s in Namespace: %s. No proxy created for this cluster.", cluster.Name, cluster.Namespace)
		}

		remote, err := url.Parse(string(secret.Data["url"]))
		if err != nil {
			log.Printf("No URL found in secret %s in Namespace: %s. No proxy created for this cluster.", cluster.Name, cluster.Namespace)
		}

		(*clusterMap)[cluster.Namespace+cluster.Name] = remote
	}
}

func clusterMapAdd(client client.Client, clusterMap *map[string]*url.URL, namespace string, name string) *url.URL {
	ref := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}

	secret := v1.Secret{}
	err := client.Get(context.TODO(), ref, &secret)
	if err != nil {
		return nil
	}

	remote, err := url.Parse(string(secret.Data["url"]))
	if err != nil {
		return nil
	}

	(*clusterMap)[namespace+name] = remote

	return (*clusterMap)[namespace+name]
}

func clusterMapRemove(clusterMap *map[string]*url.URL, namespace string, name string) {
	delete((*clusterMap), namespace+name)
}
