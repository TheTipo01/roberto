FROM --platform=$BUILDPLATFORM golang:alpine AS build

COPY go.mod /roberto/go.mod
COPY go.sum /roberto/go.sum
WORKDIR /roberto

ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go mod download

COPY . /roberto

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -trimpath -ldflags '-s -w' -o roberto

FROM alpine

RUN --mount=type=cache,target=/var/cache/apk \
    ln -s /var/cache/apk /etc/apk/cache && \
    apk add ffmpeg ca-certificates

COPY --from=thetipo01/dca /usr/bin/dca /usr/bin/
COPY --from=build /roberto/roberto /usr/bin/

CMD ["roberto"]