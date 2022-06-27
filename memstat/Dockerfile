FROM golang:1.18 as builder

WORKDIR /go/src/memstat
COPY ./ ./
RUN go build

FROM gcr.io/distroless/base
USER 1001
COPY --from=builder --chown=1001:1001 /go/src/memstat/memstat /bin/memstat

ENTRYPOINT ["/bin/memstat"]
