FROM golang:1.18 as builder

WORKDIR /go/src/logs-collector
COPY ./logs-collector ./
RUN go build

FROM bitnami/kubectl
WORKDIR /bin
COPY --from=builder /go/src/logs-collector/logs-collector ./
COPY logs-collector/increase-verbosity.sh ./
COPY logs-collector/logs-collector.sh ./
ENTRYPOINT ["/bin/logs-collector"]
