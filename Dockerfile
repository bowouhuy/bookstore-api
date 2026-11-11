# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath \
	-ldflags="-s -w -X github.com/bowouhuy/bookstore-api/config.Version=${VERSION}" \
	-o /out/bookstore-api .

# distroless: no shell or package manager, runs as a non-root user.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/bookstore-api /bookstore-api

EXPOSE 8080
ENTRYPOINT ["/bookstore-api"]
CMD ["http:start"]
