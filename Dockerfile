# Usa una imagen base de Go para construir la aplicación
FROM golang:latest AS builder

WORKDIR /app

# Copia los archivos del módulo Go
COPY ./app/go.mod ./

# Descarga las dependencias (si las hubiera, en este caso solo módulos estándar)
RUN go mod download

# Copia el código fuente de la aplicación
COPY ./app/main.go ./

# Construye la aplicación Go
# CGO_ENABLED=0 para una construcción estática sin dependencias de C (más portable)
# -o /catalog-service: nombre del ejecutable
RUN CGO_ENABLED=0 go build -o /catalog-service

# --- Segunda etapa: Imagen final de producción (más pequeña) ---
FROM alpine:latest

WORKDIR /usr/local/bin

# Copia el ejecutable desde la etapa 'builder'
COPY --from=builder /catalog-service .

# Expone el puerto en el que la aplicación Go escuchará
EXPOSE 8000

# Define el comando para ejecutar la aplicación cuando el contenedor se inicie
ENTRYPOINT ["/usr/local/bin/catalog-service"]