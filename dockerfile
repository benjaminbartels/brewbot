FROM golang:1.17 as builder

WORKDIR /src

COPY . .

RUN make build

FROM scratch as production

WORKDIR /

COPY --from=builder /src/out/bin/brewbot . 
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 1001:1001

ENTRYPOINT ["./brewbot"]