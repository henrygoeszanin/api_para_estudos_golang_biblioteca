# =========================================================
# ESTÁGIO 1: COMPILAÇÃO
# =========================================================
# Usa golang:1.24.2-alpine como imagem base para compilação
# Alpine é escolhido por ser uma imagem leve (aproximadamente 5MB)
FROM golang:1.24.2-alpine AS builder

# Instala pacotes necessários para compilar aplicações Go com CGO
# gcc: Compilador C necessário para alguns pacotes Go
# musl-dev: Biblioteca C para sistemas Alpine
# Essas libs em C são colocadas para garantir compatibilidade com libs e dependencias de libs
RUN apk add --no-cache gcc musl-dev

# Define o diretório de trabalho dentro do container
# Os comandos subsequentes serão executados neste diretório
WORKDIR /app

# Estratégia de cache: copia apenas arquivos de dependência primeiro
# Isso permite reutilizar camadas de cache se apenas o código fonte mudar
COPY go.mod go.sum ./
# Baixa todas as dependências definidas no go.mod
RUN go mod download

# Copia todo o código fonte para o container
# Este comando é executado depois da instalação das dependências para aproveitar o cache
COPY . .

# Compila a aplicação Go com otimizações:
# CGO_ENABLED=0: Desabilita CGO para gerar um binário estático
# GOOS=linux: Garante compilação para Linux independente do SO host
# -ldflags="-w -s": Remove informações de debug para reduzir o tamanho do binário
RUN CGO_ENABLED=0 GOOS=linux go build -o api_estudos_biblioteca -ldflags="-w -s" .

# =========================================================
# ESTÁGIO 2: IMAGEM FINAL
# =========================================================
# Usa Alpine como imagem base para execução - extremamente leve (aproximadamente 5MB)
FROM alpine:3.21.3

# Configura o ambiente de execução:
# - ca-certificates: Necessário para HTTPS
# - tzdata: Permite configuração de fuso horário
# - Configura o fuso horário para São Paulo/Brasil
RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/America/Sao_Paulo /etc/localtime && \
    echo "America/Sao_Paulo" > /etc/timezone

# Define o diretório de trabalho para a aplicação
WORKDIR /app

# Copia APENAS o binário compilado do estágio anterior
# Isto mantém a imagem final pequena, sem código fonte ou ferramentas de compilação
COPY --from=builder /app/api_estudos_biblioteca .

# Cria um diretório para arquivos de configuração
# Útil para montar volumes de configuração externos
RUN mkdir -p /app/config

# Documenta qual porta a aplicação vai usar
# Nota: Isso não publica a porta automaticamente, apenas documenta
EXPOSE 8080

# Define o comando que será executado quando o container iniciar
# Executa o binário da API diretamente
CMD ["./api_estudos_biblioteca"]