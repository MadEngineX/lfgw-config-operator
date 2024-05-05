# syntax=docker/dockerfile:1

FROM golang:1.20 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /lfgw-config-operator cmd/main.go

##

FROM gcr.io/distroless/base-debian11 AS runtime
WORKDIR /

COPY --from=build /lfgw-config-operator /lfgw-config-operator

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/lfgw-config-operator"]