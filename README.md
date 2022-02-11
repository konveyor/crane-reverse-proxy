# Reverse Proxy Proof of concept

# Installation
oc create -f rbac.yml
oc create -f deploy.yml
oc create -f dev-route.yml

# Basic Usage
- oc edit configmap -n openshift-migration crane-proxy
- Add the refs to cluster secrets:
```
data:
  clusters: '[{"namespace": "foo", "name": "bar"}]'
```
- Base64 encode the cluster url
```
echo -ne 'https://api.openshift.cluster.example.com:6443' | base64 -w0
```
- Create the at the namespace/name ref from the configmap.
```
apiVersion: v1
kind: Secret
metadata:
  namespace: foo
  name: bar
data:
  url: aHR0cHM6Ly9hcGkub3BlbnNoaWZ0LmNsdXN0ZXIuZXhhbXBsZS5jb206NjQ0Mw==
type: Opaque

```

A route will be created for the proxy at https://proxy-openshift-migration.cluster.base-domain.  
  
`oc get route -n openshift-migration proxy` can be used to view the URL.
  
Clusters are proxied at https://route.url/namespace/name/  
