FROM golang:1.23.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o stock-pulse ./cmd/api/main.go

# ================================================
# Segunda etapa de la construcción de la imagen
# ================================================


FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/stock-pulse .

# Instalar certificados de raíz de confianza (CA) para permitir que la aplicación se comunique de forma segura con otros servicios.
RUN apk --no-cache add ca-certificates

EXPOSE 8080

CMD ["./stock-pulse"]
