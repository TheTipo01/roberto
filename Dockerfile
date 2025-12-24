FROM --platform=$BUILDPLATFORM golang:alpine AS build

COPY . /roberto
WORKDIR /roberto

RUN apk add --no-cache ca-certificates

ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go mod download
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o roberto

FROM scratch

COPY --from=build /roberto/roberto /usr/bin/
COPY --from=build /etc/ssl/certs /etc/ssl/certs
COPY --from=build /usr/share/ca-certificates /usr/share/ca-certificates

CMD ["roberto"]