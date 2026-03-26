FROM node:24-bookworm-slim AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

FROM golang:1.24-bookworm AS backend-builder
WORKDIR /app/backend

RUN apt-get update && apt-get install -y --no-install-recommends build-essential pkg-config && rm -rf /var/lib/apt/lists/*

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

ARG BUNKO_VERSION=dev
RUN CGO_ENABLED=1 go build -ldflags="-X bunko/backend/server.Version=${BUNKO_VERSION}" -o /out/bunko .

FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

COPY --from=backend-builder /out/bunko /app/bunko
COPY backend/migrations /app/migrations
COPY --from=frontend-builder /app/frontend/dist/frontend/browser /app/public/browser

ENV BUNKO_DATABASE=/data/bunko.db
ENV BUNKO_LISTEN_ADDR=:3000

VOLUME ["/data"]
EXPOSE 3000

CMD ["/app/bunko"]
