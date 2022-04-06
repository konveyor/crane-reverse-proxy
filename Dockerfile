FROM registry.access.redhat.com/ubi8/go-toolset:latest AS builder
WORKDIR $APP_ROOT/src/github.com/jmontleon/reverse-proxy-poc
COPY . .
RUN go build -o $APP_ROOT/reverse-proxy main.go

FROM registry.redhat.io/openshift4/ose-cli:latest as manifests
COPY ./config /config
RUN kubectl kustomize /config/default > /deploy.yaml

FROM registry.access.redhat.com/ubi8-minimal
WORKDIR /
COPY --from=builder /opt/app-root/reverse-proxy .
COPY --from=manifests /deploy.yaml /deploy.yaml
ENTRYPOINT ["/reverse-proxy"]
