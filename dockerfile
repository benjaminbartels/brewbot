FROM golang:1.23 AS builder

WORKDIR /src

COPY . .

RUN make build

FROM scratch AS production

WORKDIR /

COPY --from=builder /src/out/bin/brewbot .
COPY --from=builder /src/styles.json .  
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 1001:1001

ENTRYPOINT ["./brewbot"]