FROM golang:1.18 as builder

WORKDIR /go/src/memstat
COPY memstat/ ./
RUN go build
RUN chown 1001:1001 memstat

FROM gcr.io/distroless/base
USER 1001
COPY --from=builder /go/src/memstat/memstat /bin/memstat
ENTRYPOINT ["/bin/memstat"]
