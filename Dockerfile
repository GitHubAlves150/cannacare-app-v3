# ================================================================
# CANNACARE - BACKEND DOCKERFILE
# ================================================================
# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copiar dependências
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fonte
COPY . .

# Build da aplicação
RUN go build -o cannacare-api ./cmd/api/main.go

# ================================================================
# Runtime stage
FROM alpine:latest

WORKDIR /app

# Instalar dependências de sistema
RUN apk --no-cache add ca-certificates

# Copiar binário
COPY --from=builder /app/cannacare-api .

EXPOSE 8080

CMD ["./cannacare-api"]