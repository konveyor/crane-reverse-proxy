package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	migapi "github.com/konveyor/mig-controller/pkg/apis/migration/v1alpha1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	k8sconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {
	scheme := runtime.NewScheme()

	if err := migapi.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := v1.AddToScheme(scheme); err != nil {
		panic(err)
	}

	config := k8sconfig.GetConfigOrDie()

	client, err := k8sclient.New(config, k8sclient.Options{Scheme: scheme})
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	ref := types.NamespacedName{
		Namespace: "openshift-migration",
		Name:      "migration-controller",
	}

	controller := migapi.MigrationController{}

	err = client.Get(context.TODO(), ref, &controller)
	if err != nil {
		panic(err)
	}

	for _, cluster := range controller.Spec.Clusters {
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

	r.RunTLS(":8080", os.Getenv("TLSCrt"), os.Getenv("TLSKey"))
}
