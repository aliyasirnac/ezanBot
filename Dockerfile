# 1. Adım: Go uygulamasını derlemek için kullanacağımız builder imajı
FROM golang:1.23-alpine AS builder

# Çevre değişkenlerini ayarlıyoruz (GOARCH ve GOOS çapraz derleme için)
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64


RUN apk add --no-cache git ca-certificates && update-ca-certificates

# Bağımlılıkları indirmek için çalışma dizinini ayarlıyoruz
WORKDIR /app

# Bağımlılıkları (go.mod ve go.sum) kopyalayıp indirelim
COPY go.mod go.sum ./
RUN go mod download

# Uygulama dosyalarını kopyalıyoruz
COPY . .

# Uygulamayı derliyoruz
RUN go build -o build/bot cmd/bot/main.go

# 2. Adım: Distroless imajını kullanarak yalnızca çalıştırmak için gerekli dosyaları kopyalıyoruz
FROM gcr.io/distroless/static-debian12

# SSL sertifikalarını ekliyoruz
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

# Çalışma dizinini ayarlıyoruz
WORKDIR /root/

# Derlenmiş bot dosyasını kopyalıyoruz
COPY --from=builder /app/build/bot .

# Yapılandırma dosyasını kopyalıyoruz
COPY --from=builder /app/config.yaml config.yaml

# API veya bir ağ servisi olmadığı için portu açmamıza gerek yok
# EXPOSE komutu kaldırıldı

# Uygulamayı başlatıyoruz
CMD ["./bot"]
