# Builder
FROM --platform=$BUILDPLATFORM whatwewant/builder-go:v1.22-1 as builder

WORKDIR /build

COPY go.mod ./

COPY go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
  GOOS=$TARGETOS \
  GOARCH=$TARGETARCH \
  go build \
  -trimpath \
  -ldflags '-w -s -buildid=' \
  -v -o pipeline ./cmd/pipeline

# Server
FROM whatwewant/zmicro:v1.24

LABEL MAINTAINER="Zero<tobewhatwewant@gmail.com>"

LABEL org.opencontainers.image.source="https://github.com/go-idp/pipeline"

RUN zmicro update -a

RUN apt update -y && apt install -y openssh-client vim

RUN zmicro package install docker

RUN zmicro package install docker-compose

RUN zmicro package install kubectl

RUN zmicro package install helm

RUN zmicro package install gzcaas

ARG VERSION=latest

ENV VERSION=${VERSION}

COPY --from=builder /build/pipeline /bin

CMD pipeline server
