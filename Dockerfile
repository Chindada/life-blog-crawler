# build-stage
FROM golang:1.21.3-bullseye as build-stage
USER root

ENV TZ=Asia/Taipei
ENV CLOUDFLARE_AUTH=
ENV CLOUDFLARE_ZONE=

WORKDIR /
RUN mkdir build_space
WORKDIR /build_space
COPY . .
RUN CGO_ENABLED=0 go build -o lbcrawler ./cmd/app

# production-stage
FROM debian:bullseye as production-stage
USER root

ENV TZ=Asia/Taipei

WORKDIR /
RUN apt update -y && \
    apt install -y tzdata && \
    apt autoremove -y && \
    apt clean && \
    mkdir lbcrawler && \
    mkdir lbcrawler/logs && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /lbcrawler

COPY --from=build-stage /build_space/lbcrawler ./lbcrawler

ENTRYPOINT ["/lbcrawler/lbcrawler", "--cloudflare-auth", ${CLOUDFLARE_AUTH}, "--cloudflare-zone", ${CLOUDFLARE_ZONE}]
