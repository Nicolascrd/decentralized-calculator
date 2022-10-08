FROM golang

WORKDIR /usr/src/app

EXPOSE 8000

COPY . .

RUN go build

CMD ["/usr/src/app/calculator-server"]

 
