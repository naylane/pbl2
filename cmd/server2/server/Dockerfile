FROM golang:1.24-alpine

WORKDIR /app
COPY . .

RUN ln -s /Server2/dados/empresas.json ./empresas.json

RUN go mod tidy
RUN go build -o server .
CMD ["./server"]
