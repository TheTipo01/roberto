FROM --platform=$BUILDPLATFORM golang:alpine AS build

COPY . /roberto
WORKDIR /roberto

ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go mod download
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o roberto

FROM scratch

COPY --from=build /roberto/roberto /usr/bin/

CMD ["roberto"]