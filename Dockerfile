FROM golang:1.27-trixie AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /linkforge ./cmd/api

###

FROM alpine:3.24 AS runtime

RUN adduser -D -u 10001 app

COPY --from=builder /linkforge /linkforge

USER app
EXPOSE 8080

CMD [ "/linkforge" ]
