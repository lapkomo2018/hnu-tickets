FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

FROM alpine:3.22
RUN adduser -D -H -u 10001 app
USER app
COPY --from=build /api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
