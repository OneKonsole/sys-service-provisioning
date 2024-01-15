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
ENV RABBITMQ_USER=null
ENV RABBITMQ_PASSWORD=null
ENV RABBITMQ_HOST=null
ENV RABBITMQ_QUEUE=null
ENV RABBITMQ_VHOST=null
CMD ["/sys-service-provisioning", "-t", "kubeConfig", "-d", "", "-e", "", "-s", "default"]