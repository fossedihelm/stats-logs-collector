FROM golang:1.18 as builder

WORKDIR /go/src/logs-collector
COPY ./ ./
RUN go build

FROM bitnami/kubectl
WORKDIR /bin
COPY --from=builder /go/src/logs-collector/logs-collector ./
COPY ./increase-verbosity.sh ./
COPY ./logs-collector.sh ./
ENTRYPOINT ["/bin/logs-collector"]
