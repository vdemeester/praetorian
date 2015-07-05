## -*- docker-image-name: "vdemeester/praetorian-build" -*-
FROM golang:1.4.2
MAINTAINER Vincent Demeester <vincent@sbr.pm>

# Install gb
RUN go get github.com/constabulary/gb/...

# Copy project inside
COPY . /usr/src/praetorian
WORKDIR /usr/src/praetorian

# Build it
RUN gb build
# Run tests
RUN gb test

CMD ["/usr/src/praetorian/bin/praetorian"]
