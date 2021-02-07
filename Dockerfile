FROM golang:alpine AS builder
WORKDIR /app
ADD go.mod go.mod
ADD go.sum go.sum
ADD main.go main.go
RUN go build

FROM alpine
COPY --from=builder /app/url-monitor /url-monitor
ENTRYPOINT ["/url-monitor"]

# a valid configuration.yaml needs to be mounted into the container at /configuration.yaml
CMD [ "--configuration-file", "/configuration.yaml"]