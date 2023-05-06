FROM golang:1.20 as build

WORKDIR /go/src/practice-4
COPY . .

RUN go test ./...
ENV CGO_ENABLED=0
RUN go install ./cmd/...

# ==== Final image ====
FROM alpine:latest
WORKDIR /opt/practice-4
COPY entry.sh /opt/practice-4/
COPY --from=build /go/bin/* /opt/practice-4
RUN ls /opt/practice-4
ENTRYPOINT ["/opt/practice-4/entry.sh"]
CMD ["server"]
