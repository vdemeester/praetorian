FROM golang:1.6

RUN apt-get update && apt-get install -y \
    build-essential \
    --no-install-recommends

# Install build dependencies
RUN go get golang.org/x/tools/cmd/cover \
    && go get github.com/golang/lint/golint

# Which docker version to test on
ENV DOCKER_VERSION 1.10.3

# Download docker
RUN set -ex; \
    curl https://get.docker.com/builds/Linux/x86_64/docker-${DOCKER_VERSION} -o /usr/local/bin/docker; \
    chmod +x /usr/local/bin/docker
    
WORKDIR /go/src/github.com/vdemeester/praetorian

COPY . /go/src/github.com/vdemeester/praetorian
