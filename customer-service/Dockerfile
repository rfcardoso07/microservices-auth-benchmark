FROM golang:1.18

COPY . /app

WORKDIR /app

RUN go mod vendor

EXPOSE 8000

CMD ["go", "run", "customer.go"]