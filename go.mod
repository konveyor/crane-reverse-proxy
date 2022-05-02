module github.com/jmontleon/reverse-proxy-poc

go 1.16

require (
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.7.7
	github.com/go-playground/validator/v10 v10.8.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.19.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/ugorji/go v1.2.6 // indirect
	k8s.io/api v0.23.5
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v0.23.5
	sigs.k8s.io/controller-runtime v0.11.2
)
