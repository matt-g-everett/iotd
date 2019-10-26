FROM golang:1.13.3-buster as builder

#RUN adduser --no-create-home --disabled-password --uid 1 --gid 1 daemon

WORKDIR $GOPATH/src/github.com/matt-g-everett/iotd/

COPY . .

RUN go get -d -v

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -a -installsuffix cgo -o /go/bin/iotd .


FROM scratch

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/iotd /go/bin/iotd

USER daemon

ENTRYPOINT ["/go/bin/iotd"]
