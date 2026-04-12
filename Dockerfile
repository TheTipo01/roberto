FROM golang:trixie AS build

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/lists,sharing=locked \
    apt-get update && apt-get install unzip -y

COPY . /roberto
WORKDIR /roberto

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

RUN wget https://raw.githubusercontent.com/disgoorg/godave/refs/heads/master/scripts/libdave_install.sh && chmod +x libdave_install.sh
ENV SHELL=/bin/sh
RUN ./libdave_install.sh v1.1.0

ENV PKG_CONFIG_PATH="/root/.local/lib/pkgconfig"
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags "-s -w" -o roberto

FROM debian:trixie-slim

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/lists,sharing=locked \
    apt-get update && apt-get install ffmpeg ca-certificates -y --no-install-recommends

COPY --from=thetipo01/dca /usr/bin/dca /usr/bin/
COPY --from=build /roberto/roberto /usr/bin/

COPY --from=build /root/.local/lib /root/.local/lib
ENV PKG_CONFIG_PATH="/root/.local/lib/pkgconfig"

CMD ["roberto"]