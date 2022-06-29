FROM golang:alpine AS build-env
# The GOPATH in the image is /go.
ADD . /go/src/melody
WORKDIR /go/src/melody/cmd/db-manager
RUN if [ "$(uname -m)" = "ppc64le" ] || [ "$(uname -m)" = "aarch64" ]; then \
        apk --update add git gcc musl-dev && \
        go build -o melody-storage .; \
    else \
        go build -o melody-storage .; \
    fi
RUN GRPC_HEALTH_PROBE_VERSION=v0.3.1 && \
    if [ "$(uname -m)" = "ppc64le" ]; then \
	wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-ppc64le; \
    elif [ "$(uname -m)" = "aarch64" ]; then \
	wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-arm64; \
    else \
	wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64; \
    fi && \
    chmod +x /bin/grpc_health_probe

FROM alpine:3.7
WORKDIR /app
COPY --from=build-env /bin/grpc_health_probe /bin/
COPY --from=build-env /go/src/melody/cmd/db-manager/melody-storage /app/
ENTRYPOINT ["./melody-storage"]
CMD ["-w", "kubernetes"]
