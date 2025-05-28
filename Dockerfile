FROM golang:1.23-bookworm AS builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Use modules cache to speed up the build process on subsequent builds on the same machine.
ENV GOMODCACHE=/root/.kube-yaml-to-go-cache/go-modules
RUN --mount=type=cache,target="/root/.kube-yaml-to-go-cache" go mod download

# Copy the go source
COPY main.go .
COPY cmd/ cmd/
COPY pkg/ pkg/

# Use build cache to speed up the build process on subsequent builds on the same machine
RUN --mount=type=cache,target="/root/.kube-yaml-to-go-cache" CGO_ENABLED=0 \
    GOOS=linux GOARCH=amd64 go build -o kube-yaml-to-go

# Use distroless as minimal base image to package the binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/kube-yaml-to-go .
USER 65532:65532

ENTRYPOINT ["/kube-yaml-to-go"]