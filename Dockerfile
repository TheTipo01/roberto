FROM golang:alpine AS build

RUN --mount=type=cache,target=/var/cache/apk \
    ln -s /var/cache/apk /etc/apk/cache && \
    apk add --no-cache build-base pkgconfig ccache

COPY go.mod /roberto/go.mod
COPY go.sum /roberto/go.sum
WORKDIR /roberto

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . /roberto

COPY --from=ghcr.io/thetipo01/godave-musl:latest /root/.local /root/.local
ENV PKG_CONFIG_PATH="/root/.local/lib/pkgconfig"

ENV CC=/usr/local/bin/gcc CXX=/usr/local/bin/g++
RUN ln -s /usr/bin/ccache /usr/local/bin/gcc && ln -s /usr/bin/ccache /usr/local/bin/g++ && ln -s /usr/bin/ccache /usr/local/bin/cc && ln -s /usr/bin/ccache /usr/local/bin/c++
ENV CCACHE_DIR=/ccache

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/ccache \
    go build -trimpath -ldflags '-s -w' -o roberto

FROM alpine

RUN --mount=type=cache,target=/var/cache/apk \
    ln -s /var/cache/apk /etc/apk/cache && \
    apk add ffmpeg ca-certificates

COPY --from=thetipo01/dca /usr/bin/dca /usr/bin/
COPY --from=build /roberto/roberto /usr/bin/

COPY --from=build /root/.local/lib /root/.local/lib
ENV PKG_CONFIG_PATH="/root/.local/lib/pkgconfig"

CMD ["roberto"]