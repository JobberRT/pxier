FROM golang:1.18

WORKDIR /pixer

COPY . .

ENV GOPROXY "https://goproxy.cn"
RUN go mod vendor
RUN go build -o pixer
RUN cp config.example.yaml config.yaml

CMD ["./pixer"]