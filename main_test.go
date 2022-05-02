package main

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/patrickmn/go-cache"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCart(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Crane Reverse Proxy Suite")
}

var _ = Describe("Crane Reverse Proxy", func() {

	Context("Initial Cache", func() {
		gocache := cache.New(5*time.Minute, 10*time.Minute)
		//	gocache.Set("sweet", "potato", cache.DefaultExpiration)

		It("cache has 0 items", func() {
			Expect(gocache.Items()).Should(BeEmpty())
		})
	})

	Context("Add a URL from a secret", func() {
		fakescheme := scheme.Scheme
		client := fake.NewClientBuilder().WithScheme(fakescheme).WithObjects(&v1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster-coordinates",
				Namespace: "secret-namespace",
			},
			Data: map[string][]byte{
				"url": []byte("https://onefakecluster:8443"),
			},
		}).Build()

		gocache := cache.New(5*time.Minute, 10*time.Minute)

		url := getClusterURL(client, gocache, "secret-namespace", "cluster-coordinates")

		It("URL Host and scheme match", func() {
			Expect(url.Host).Should(BeEquivalentTo("onefakecluster:8443"))
			Expect(url.Scheme).Should(BeEquivalentTo("https"))
		})
	})

	Context("Return nil if no secret exists", func() {
		fakescheme := scheme.Scheme
		client := fake.NewClientBuilder().WithScheme(fakescheme).WithObjects().Build()

		gocache := cache.New(5*time.Minute, 10*time.Minute)

		url := getClusterURL(client, gocache, "secret-namespace", "cluster-coordinates")

		It("URL is Nil", func() {
			Expect(url).Should(BeNil())
		})
	})

})
