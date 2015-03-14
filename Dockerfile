## -*- docker-image-name: "praetorian" -*-
FROM ruby
MAINTAINER Vincent Demeester <vincent@sbr.pm>

# Install fpm, ronn and stuff
RUN gem install ronn fpm

COPY . /usr/src

WORKDIR /usr/src

RUN ["make", "-f", "docker.Makefile", "build"]

