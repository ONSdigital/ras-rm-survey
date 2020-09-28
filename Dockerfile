FROM golang:1.15.2-alpine3.12

RUN mkdir /opt
WORKDIR /opt
COPY . .
RUN make build
CMD [ "./main" ]