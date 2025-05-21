FROM golang:1.24-alpine

WORKDIR /app
COPY . .

RUN ln -s /servidor3/dados/empresas.json ./empresas.json


RUN go mod tidy
RUN go build -o servidor .
CMD ["./servidor"]
