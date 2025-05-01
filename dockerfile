# =================================================================================
# ESTÁGIO 1: COMPILAÇÃO
# =================================================================================
# Usa Alpine como base por ser extremamente leve (~5MB vs ~300MB de imagens Debian)
# A versão específica (1.24.2) garante reprodutibilidade do build
FROM golang:1.24.2-alpine3.21 AS builder

# Instala compilador C e bibliotecas de desenvolvimento
# --no-cache: Evita armazenar os índices de pacotes APK, economizando ~1MB
# gcc: Necessário para alguns pacotes Go que usam cgo implicitamente
# musl-dev: Biblioteca C para Alpine, usada pelo gcc
# tzdata: Contém arquivos de fusos horários necessários para a aplicação
RUN apk add --no-cache gcc musl-dev tzdata

# Define variáveis que serão incorporadas no binário final
# VERSION: Identifica a versão do software (passada no build ou usa "dev")
# BUILD_TIME: Registra quando o build foi executado
# GIT_COMMIT: Rastreabilidade para identificar exatamente qual código está rodando
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

# Configurações de otimização para o compilador Go
# GOOS=linux: Garante compilação para Linux independentemente do host
# CGO_ENABLED=0: Desabilita CGO para gerar binário estático (sem dependências externas)
# GOMEMLIMIT=1024MiB: Limita uso de memória durante compilação (evita OOM em CI/CD)
# GOGC=off: Desativa garbage collector durante build para melhor performance
ENV GOOS=linux
ENV CGO_ENABLED=0
ENV GOMEMLIMIT=1024MiB
ENV GOGC=off

# Define diretório de trabalho onde os comandos serão executados
WORKDIR /app

# Estratégia de otimização de cache em camadas:
# 1. Copia apenas arquivos de dependências primeiro
# 2. Baixa dependências (esta camada só muda quando go.mod/go.sum mudam)
# 3. Reaproveita cache se apenas o código fonte mudar
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o código fonte e compila a aplicação
# -trimpath: Remove caminhos do sistema de build do binário (reprodutibilidade/segurança)
# -ldflags: Passa opções ao linker:
#   -w: Remove informações de debug DWARF (reduz tamanho)
#   -s: Remove tabela de símbolos (reduz tamanho)
#   -X: Define variáveis em tempo de compilação para inserir metadados
COPY . .
RUN go build -trimpath \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}" \
    -o api_estudos_biblioteca .


# =================================================================================
# ESTÁGIO 2: IMAGEM FINAL MÍNIMA
# =================================================================================
# "scratch" é uma imagem especial totalmente vazia (0 bytes)
# Reduz dramaticamente o tamanho final e superfície de ataque (não há shell, utilitários, etc.)
FROM scratch

# Define variáveis que serão incorporadas no binário final
# VERSION: Identifica a versão do software (passada no build ou usa "dev")
# BUILD_TIME: Registra quando o build foi executado
# GIT_COMMIT: Rastreabilidade para identificar exatamente qual código está rodando
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

# Copia certificados CA e dados de timezone da imagem builder
# Necessários para conexões HTTPS e manipulação correta de fusos horários
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=America/Sao_Paulo

# Configuração de segurança: cria usuário não-root
# Prática de segurança essencial: nunca execute como root em containers
COPY --from=builder /etc/passwd /etc/passwd
USER nobody

# Preparação para execução do aplicativo
WORKDIR /app
COPY --from=builder /app/api_estudos_biblioteca .

# Metadados da imagem (para documentação e rastreabilidade)
LABEL maintainer="Seu Nome <henrygoeszaninpro@gmail.com>"
LABEL version="${VERSION}"
LABEL description="API de Biblioteca em Golang"

# Monitoramento automático de saúde do container
# --interval=30s: Verifica a cada 30 segundos
# --timeout=3s: Considera falha após 3 segundos sem resposta
# --start-period=5s: Aguarda 5s após inicialização para começar verificações
# Usa wget para verificar endpoint /api/health e falha se não retornar HTTP 200
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD wget -q -O- http://localhost:8080/api/health || exit 1

# Documenta que a aplicação usa a porta 8080 (apenas informativo)
EXPOSE 8080

# Configuração de inicialização em duas partes:
# ENTRYPOINT: Define o executável primário que sempre roda (mais seguro)
# CMD: Argumentos padrão vazios que podem ser sobrescritos na linha de comando
# Esta abordagem permite flexibilidade na execução sem comprometer a segurança
ENTRYPOINT ["./api_estudos_biblioteca"]
CMD []