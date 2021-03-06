FROM golang:1.18 AS build
WORKDIR /pxier
COPY . .
RUN go mod tidy &&  \
    go mod vendor && \
    go build -o pxier && \
    cp config.example.yaml config.yaml

FROM ubuntu:22.04 AS run
COPY --from=build /pxier/pxier .
COPY --from=build /pxier/config.yaml .
CMD ["./pxier"]