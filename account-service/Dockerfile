FROM golang:1.18

COPY . /app

WORKDIR /app

RUN go mod vendor

EXPOSE 8001

CMD ["go", "run", "account.go"]