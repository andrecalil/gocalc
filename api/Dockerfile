# syntax=docker/dockerfile:1.7

# ---- Build stage --------------------------------------------------------
FROM golang:1.23-alpine AS build

WORKDIR /src

# Fetch dependencies in a separate layer so they're cached between builds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary — no CGO, no dynamic loader — so the runtime image can be
# distroless/static with no libc.
RUN CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w" \
      -o /out/server \
      ./cmd/server

# ---- Runtime stage ------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/server /server

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
