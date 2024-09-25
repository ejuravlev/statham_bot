FROM golang:1.22.1 AS build-stage
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o statham


FROM gcr.io/distroless/base-debian11 AS build-release-stage
WORKDIR /
COPY --from=build-stage /app/statham statham
USER nonroot:nonroot
ENTRYPOINT ["/statham"]