FROM golang:1.24-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/logmesh-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/logmesh-api /logmesh-api
EXPOSE 8080
ENTRYPOINT ["/logmesh-api"]
