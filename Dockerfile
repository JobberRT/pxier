FROM golang:1.18 AS build
WORKDIR /pixer
COPY . .
RUN go mod vendor
RUN go build -o pixer
RUN cp config.example.yaml config.yaml

FROM ubuntu:latest AS run
COPY --from=build /pixer/pixer .
COPY --from=build /pixer/config.yaml .
CMD ["./pixer"]