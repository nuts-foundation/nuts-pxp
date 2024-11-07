# golang alpine
FROM golang:1.23.3-alpine AS builder

ARG TARGETARCH
ARG TARGETOS

LABEL maintainer="wout.slakhorst@nuts.nl"

RUN apk update \
 && apk add --no-cache \
            gcc \
            musl-dev \
 && update-ca-certificates

ENV GO111MODULE=on
ENV GOPATH=/

RUN mkdir /opt/nuts-pxp && cd /opt/nuts-pxp
COPY go.mod .
COPY go.sum .
RUN go mod download && go mod verify

COPY . .
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-w -s" -o /opt/nuts/nuts-pxp

# alpine
FROM alpine:3.20.2
RUN apk update \
  && apk add --no-cache \
             tzdata \
             curl \
  && update-ca-certificates
COPY --from=builder /opt/nuts/nuts-pxp /usr/bin/nuts-pxp

RUN adduser -D -H -u 18081 nuts-usr
USER 18081:18081
WORKDIR /nuts

EXPOSE 8080
ENTRYPOINT ["/usr/bin/nuts-pxp"]
CMD []
