module github.com/jmontleon/reverse-proxy-poc

go 1.16

require (
	github.com/gin-gonic/gin v1.7.7
	github.com/konveyor/mig-controller v0.0.0-20220110144829-0b1d57511b4c
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.22.1
	sigs.k8s.io/controller-runtime v0.9.2
)

replace bitbucket.org/ww/goautoneg v0.0.0-20120707110453-75cd24fc2f2c => github.com/markusthoemmes/goautoneg v0.0.0-20190713162725-c6008fefa5b1

replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2

replace k8s.io/client-go => k8s.io/client-go v0.20.7

replace k8s.io/apimachinery => k8s.io/apimachinery v0.20.7

replace k8s.io/api => k8s.io/api v0.20.7

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.7

replace k8s.io/apiserver => k8s.io/apiserver v0.20.7

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.7

replace sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.7.1-0.20201215171748-096b2e07c091

replace github.com/konveyor/mig-controller => /home/jason/Documents/src/github.com/konveyor/mig-controller
