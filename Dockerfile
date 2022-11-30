# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# 坑：
# 报错 go mod download: google.golang.org/api@v0.44.0: read tcp 172.17.0.3:60862->14.204.51.154:443: read: connection reset by peer
# The command '/bin/sh -c go mod download' returned a non-zero code: 1
# make: *** [docker-build] 错误 1
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
# # 需要把该放入的都copy进去，如果报出 package xxxxx is not in GOROOT  => 就是这个问题。
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# 坑：
# 报错 ---> f1b27b46feba
# Step 12/16 : FROM gcr.io/distroless/static:nonroot
# Get "https://gcr.io/v2/": net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers)
# make: *** [docker-build] 错误 1
# 解决方案：
# docker pull exploitht/operator-static
# docker tag exploitht/operator-static gcr.io/distroless/static:nonroot
# docker rmi exploitht/operator-static
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]

# 坑：
# 部署后有问题：镜像问题
# Normal   Started    17s                kubelet            Started container manager
# Normal   BackOff    15s (x2 over 16s)  kubelet            Back-off pulling image "gcr.io/kubebuilder/kube-rbac-proxy:v0.11.0"
# Warning  Failed     15s (x2 over 16s)  kubelet            Error: ImagePullBackOff
# Normal   Pulling    4s (x2 over 32s)   kubelet            Pulling image "gcr.io/kubebuilder/kube-rbac-proxy:v0.11.0"

# 解决方案：
# docker pull kubesphere/kube-rbac-proxy:v0.8.0
# docker tag kubesphere/kube-rbac-proxy:v0.8.0 gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0
# docker rmi kubesphere/kube-rbac-proxy:v0.8.0