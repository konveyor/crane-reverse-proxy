package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	proxySecretName = "crane-proxy"
)

type Cluster struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Unable to retrieve in cluster kubeconfig.")
	}

	gocache := cache.New(5*time.Minute, 10*time.Minute)

	client, err := client.New(config, client.Options{})
	if err != nil {
		log.Fatalf("Unable to create kubernetes client.")
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Any("/:namespace/:name/*proxyPath", func(c *gin.Context) {
		var proxy *httputil.ReverseProxy

		namespace, _ := c.Params.Get("namespace")
		name, _ := c.Params.Get("name")

		url := getClusterURL(client, gocache, namespace, name)
		if url != nil {
			proxy = httputil.NewSingleHostReverseProxy(url)
		}

		if proxy == nil {
			c.AbortWithStatus(http.StatusBadGateway)
		} else {
			proxy.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			proxy.FlushInterval = 0
			c.Request.URL.Path, _ = c.Params.Get("proxyPath")
			c.Request.Host = url.Host
			c.Request.Header.Del("origin")
			proxy.ServeHTTP(c.Writer, c.Request)

			if c.Writer.Status() >= http.StatusBadRequest {
				gocache.Delete(namespace + name)
			}

		}
	})

	crt := os.Getenv("CRANE_PROXY_CRT")
	key := os.Getenv("CRANE_PROXY_KEY")
	if crt == "" || key == "" {
		log.Fatalf("Export CRANE_PROXY_CRT and CRANE_PROXY_KEY before running.")
	}

	r.RunTLS(":8443", crt, key)
}

func getClusterURL(client client.Client, gocache *cache.Cache, namespace string, name string) *url.URL {
	cachedRemote, found := gocache.Get(namespace + name)
	if found {
		return cachedRemote.(*url.URL)
	}

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

	gocache.Set(namespace+name, remote, cache.DefaultExpiration)
	return remote
}
