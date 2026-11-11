# syntax=docker/dockerfile:1

# --- stage 1: TypeScript compile (no bundler, plain ES modules) -------------
FROM node:24-alpine AS web
WORKDIR /app
COPY package.json ./
RUN npm install --no-audit --no-fund --no-progress
COPY tsconfig.json ./
COPY web/ts ./web/ts
RUN npx tsc -p tsconfig.json

# --- stage 2: Go build (static, CGO off) -----------------------------------
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /app/web/static/js ./web/static/js
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# --- stage 3: runtime (distroless, ~12 MB) ---------------------------------
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server
ENV ADDR=:8080 DATA_DIR=/data
EXPOSE 8080
ENTRYPOINT ["/server"]