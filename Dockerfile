## -*- docker-image-name: "vdemeester/praetorian-build" -*-
FROM golang:1.6
MAINTAINER Vincent Demeester <vincent@sbr.pm>

# Install gb
RUN go get github.com/constabulary/gb/... && \
    go get github.com/golang/lint/golint && \
    go get golang.org/x/tools/cmd/cover && \
    go get github.com/modocache/gover && \
    go get github.com/mattn/goveralls

# Copy project inside
COPY . /usr/src/praetorian
WORKDIR /usr/src/praetorian
