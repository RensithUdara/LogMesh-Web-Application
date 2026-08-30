FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/logmesh-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/logmesh-processor ./cmd/processor

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/logmesh-api /logmesh-api
COPY --from=build /out/logmesh-processor /logmesh-processor
EXPOSE 8081
CMD ["/logmesh-api"]
