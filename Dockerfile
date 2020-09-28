FROM golang:1.15.2-alpine3.12
RUN apk add --no-cache make

WORKDIR /opt
COPY . .
RUN make build
CMD [ "./main" ]