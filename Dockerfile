# build stage
FROM --platform=$BUILDPLATFORM golang:1.14-stretch AS build-env

ADD . /src
ENV CGO_ENABLED=0
WORKDIR /src

ARG TARGETOS
ARG TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o rpi_exporter

# final stage
FROM scratch
WORKDIR /app
COPY --from=build-env /src/rpi_exporter /app/
ENTRYPOINT ["/app/rpi_exporter"]
