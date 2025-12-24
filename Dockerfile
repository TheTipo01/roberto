FROM --platform=$BUILDPLATFORM golang:alpine AS build

COPY . /roberto
WORKDIR /roberto

RUN apk add --no-cache ca-certificates

ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go mod download
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o roberto

FROM alpine

RUN apk add --no-cache ca-certificates ffmpeg

COPY --from=build /roberto/roberto /usr/bin/
COPY --from=thetipo01/dca /usr/bin/dca /usr/bin/

CMD ["roberto"]