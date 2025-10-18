FROM golang:1.25-alpine AS builder

WORKDIR /go/src/app

RUN apk update && apk add upx ca-certificates tzdata

ARG VERSION=main
ARG BUILD="N/A"

ENV GO111MODULE=on \
  CGO_ENABLED=0

COPY . /go/src/app/

RUN go build -a -installsuffix cgo -ldflags="-w -s -X github.com/peak-scale/break-the-glass/internal/version.Version=${VERSION} -X github.com/peak-scale/break-the-glass/internal/version.Build=${BUILD}" -o break-the-glass ./cmd/

RUN go version && upx -q break-the-glass

# application image
FROM scratch
WORKDIR /opt/go

LABEL maintainer="TBD"
EXPOSE 8080
ENTRYPOINT ["/opt/go/break-the-glass"]
CMD ["run", "--config", "/config/break-the-glass.yaml"]
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/ /usr/share/zoneinfo/
COPY --from=builder /go/src/app/break-the-glass  /opt/go/break-the-glass
USER 1001
