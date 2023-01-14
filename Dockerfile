# Builder image.
FROM golang as builder

COPY . /build
WORKDIR /build

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -o ./ethService ./app/main.go

# Runtime image.
FROM alpine

ENV TZ=Asia/Taipei

RUN apk --update-cache add ca-certificates tzdata
COPY --from=builder /build/ethService /

EXPOSE 8080

ENTRYPOINT [ "/ethService" ]

