FROM golang:1.25 AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o turnify .

FROM scratch
COPY --from=build /app/turnify /turnify
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
USER 65532
ENTRYPOINT ["/turnify"]
EXPOSE 4499

LABEL org.opencontainers.image.title="turnify"
LABEL org.opencontainers.image.description="Middleman proxy service to allow usage of Cloudflare TURN server with Matrix"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.url="https://github.com/mister-mr-matrix/turnify"
LABEL org.opencontainers.image.source="https://github.com/mister-mr-matrix/turnify"
LABEL org.opencontainers.image.documentation="https://github.com/mister-mr-matrix/turnify#readme"
