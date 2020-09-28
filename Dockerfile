FROM golang:1.15.2-alpine3.12


WORKDIR /opt
COPY . .
RUN make build
CMD [ "./main" ]