FROM golang:1.20-alpine as build
WORKDIR /go/src/godis
COPY . .
RUN CGO_ENABLED=0
RUN go build -o /go/bin/godis ./cmd/godis/main.go

FROM scratch
COPY --from=build /go/bin/godis /bin/godis
ENTRYPOINT [ "/bin/godis" ]