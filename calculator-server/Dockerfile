FROM golang

WORKDIR /usr/src/app

ARG PORT 

EXPOSE $PORT

COPY . .

RUN go build

CMD ["/usr/src/app/calculator-server"]

 
