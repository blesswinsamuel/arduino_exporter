# build stage
FROM --platform=$BUILDPLATFORM golang:1.14-stretch AS build-env

ADD . /src
ENV CGO_ENABLED=0
WORKDIR /src

ARG TARGETOS
ARG TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o arduino_exporter

# final stage
FROM scratch
WORKDIR /app
COPY --from=build-env /src/arduino_exporter /app/
ENTRYPOINT ["/app/arduino_exporter"]
