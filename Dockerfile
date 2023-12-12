FROM --platform=linux/amd64 golang:1.21.5 AS build

WORKDIR /go/bin/app

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/app

FROM gcr.io/distroless/static-debian11:nonroot
WORKDIR /
COPY --from=build /go/bin/app .
USER 65532:65532
ENTRYPOINT ["/sys-service-provisioning"]