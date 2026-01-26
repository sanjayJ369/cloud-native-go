FROM golang:1.25.6-alpine AS builder

WORKDIR /app 

COPY . .

RUN go build -o main ./cmd/server*

FROM alpine:latest AS runner 

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 80

CMD [ "./main" ]





