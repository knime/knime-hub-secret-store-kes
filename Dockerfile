###########
# sources
###########
# golang 1.24.4 for multi-platform
# https://hub.docker.com/layers/library/golang/1.24.4/images/sha256-be70d93633d07a2acae4ff3401672b04f23e5850b0248d65c23e30dc75dded09
FROM golang:1.24.4@sha256:10c131810f80a4802c49cab0961bbe18a16f4bb2fb99ef16deaa23e4246fc817 AS sources

# Copy deps (vendor).
# see: https://blog.boot.dev/golang/should-you-commit-the-vendor-folder-in-go/
WORKDIR /builddir
COPY go.mod ./
COPY go.sum ./
# COPY vendor ./vendor/
COPY .git ./.git/

# Copy (actual) sources.
COPY *.go ./
COPY internal ./internal/
COPY cmd ./cmd/
COPY kesconf ./kesconf/


###########
# test
###########
# Lint and test, create SBOM, and export results.
FROM sources AS test

# Copy golangci config file.
COPY .golangci.yml ./

# Install golangci-lint.
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/v1.64.8/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.8

# Lint.
RUN golangci-lint run  ./... --out-format checkstyle:/tmp/lint.out,colored-line-number

# Install the got test to junit converter.
RUN go install github.com/jstemmer/go-junit-report/v2@latest

# Run tests.
RUN go test ./... -p=1 -coverpkg=./... -coverprofile /tmp/cover.out -v | ${GOPATH}/bin/go-junit-report > /tmp/report.xml

# Install and run the SBOM creator.
RUN go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.9.0
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ${GOPATH}/bin/cyclonedx-gomod app -json -output /tmp/bom.json -licenses --assert-licenses=true -main ./cmd/kes .


###########
# export-test-results
###########
# Export lint- and test-results and SBOM.
# Use scratch for exporting the test results.
# see: https://kevsoft.net/2021/08/09/exporting-unit-test-results-from-a-multi-stage-docker-build.html
FROM scratch AS export-test-results
COPY --from=test /tmp/report.xml .
COPY --from=test /tmp/lint.out .
COPY --from=test /tmp/cover.out .
COPY --from=test /tmp/bom.json .

###########
# build
###########
FROM sources AS build

# Build the application / service.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-w -s" \
    -o kes \
    ./cmd/kes

# Create app user.
# see https://stackoverflow.com/a/55757473/12429735RUN
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "10001" \
    "app"

###########
# base
###########
# To allow tracking base-images independently from applications (in DTrack),
# we create (on-the-fly) a "base image" (i.e. scratch plus everything BUT the
# application itself).
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.6 AS ubi-minimal

RUN microdnf update -y --nodocs && microdnf install ca-certificates --nodocs

FROM registry.access.redhat.com/ubi9/ubi-micro:9.6 AS base

# On RHEL the certificate bundle is located at:
# - /etc/pki/tls/certs/ca-bundle.crt (RHEL 6)
# - /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem (RHEL 7)
COPY --from=ubi-minimal /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem /etc/pki/ca-trust/extracted/pem/

# Copy the user- and group files from the builder stage.
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group

###########
# final image
###########
# Use the above created base-image (ubi-micro) the for the final image.
FROM base

ARG TAG

#LABEL name="MinIO" \
#      vendor="MinIO Inc <dev@min.io>" \
#      maintainer="MinIO Inc <dev@min.io>" \
#      version="${TAG}" \
#      release="${TAG}" \
#      summary="KES is a cloud-native distributed key management and encryption server designed to build zero-trust infrastructures at scale."

COPY LICENSE /LICENSE
COPY CREDITS /CREDITS
COPY --from=build /builddir/kes /kes

EXPOSE 7373

# Use app:app as (no-root) user.
USER app:app

ENTRYPOINT ["/kes"]
CMD ["kes"]
