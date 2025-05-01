# API de Biblioteca em Golang

Uma API RESTful completa para gerenciamento de biblioteca, desenvolvida com Go e seguindo princípios de Clean Architecture e Domain-Driven Design (DDD).

## 📋 Descrição

Este projeto implementa uma API para gerenciamento de biblioteca que permite:

- Cadastro e autenticação de usuários
- CRUD completo de livros (apenas para administradores)
- Empréstimos de livros por períodos definidos
- Devolução de livros
- Gerenciamento de usuários e permissões

O primeiro usuário cadastrado no sistema é automaticamente definido como administrador, e apenas administradores podem promover outros usuários a administradores.

## 🛠️ Tecnologias

- **Go 1.24+**: Linguagem de programação principal
- **Gin**: Framework web para construção de APIs
- **GORM**: ORM para Go com PostgreSQL
- **gin-jwt**: Middleware para autenticação JWT
- **PostgreSQL**: Banco de dados relacional
- **godotenv**: Carregamento de variáveis de ambiente

## 🏗️ Arquitetura

O projeto segue os princípios de Clean Architecture e Domain-Driven Design, com a seguinte estrutura:

```
api_golang_estudos/
├── application/                # Camada de aplicação
│   ├── services/               # Implementação da lógica de negócios
│   ├── dtos/                   # Objetos de transferência de dados
│   └── interfaces/             # Interfaces de repositórios
├── config/                     # Configurações da aplicação
├── domain/                     # Camada de domínio
│   ├── entities/               # Entidades do domínio
│   └── errors/                 # Erros específicos do domínio
├── infrastructure/             # Camada de infraestrutura
│   ├── database/               # Configuração do banco de dados
│   └── repositories/           # Implementação dos repositórios
├── presentation/               # Camada de apresentação
│   ├── handlers/               # Manipuladores de requisições HTTP
│   ├── middlewares/            # Middlewares da aplicação
│   └── routes/                 # Definição de rotas da API
└── main.go                     # Ponto de entrada da aplicação
```

## ⚙️ Configuração e Execução

### Pré-requisitos

- Go 1.24 ou superior
- PostgreSQL
- Docker (opcional, para execução do PostgreSQL)

### Variáveis de ambiente

Crie um arquivo `.env` na raiz do projeto baseado no `.env.example`:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=library_api
SERVER_PORT=8080
JWT_SECRET=chave_secreta_muito_segura_aqui
```

### Instalação

Clone o repositório:

```sh
git clone https://github.com/henrygoeszanin/api_golang_estudos.git
cd api_golang_estudos
```

Instale as dependências:

```sh
go mod tidy
```

Inicie o PostgreSQL:

```sh
docker run --name postgres-library -e POSTGRES_PASSWORD=postgres -e POSTGRES_USER=postgres -e POSTGRES_DB=library_api -p 5432:5432 -d postgres
```

Execute a aplicação:

```sh
go run main.go
```

## 🔀 Endpoints da API

### Autenticação

- `POST /api/auth/register`: Registrar novo usuário
- `POST /api/auth/login`: Autenticar usuário
- `GET /api/auth/refresh`: Renovar token JWT

### Usuários

- `GET /api/users/me`: Obter dados do usuário atual
- `PUT /api/users/me`: Atualizar dados do usuário atual

#### Rotas Administrativas (requer permissão de administrador)

- `GET /api/admin/users`: Listar todos os usuários
- `GET /api/admin/users/:id`: Obter usuário específico
- `PUT /api/admin/users/:id`: Atualizar usuário
- `DELETE /api/admin/users/:id`: Remover usuário
- `PUT /api/admin/users/:id/promote`: Promover usuário para administrador

### Livros

- `GET /api/books`: Listar todos os livros
- `GET /api/books/:id`: Obter livro específico

#### Rotas Administrativas (requer permissão de administrador)

- `POST /api/admin/books`: Adicionar novo livro
- `PUT /api/admin/books/:id`: Atualizar livro
- `DELETE /api/admin/books/:id`: Remover livro

### Empréstimos (requer autenticação)

- `GET /api/loans`: Listar empréstimos do usuário atual
- `POST /api/loans`: Criar novo empréstimo
- `GET /api/loans/:id`: Obter empréstimo específico
- `PUT /api/loans/:id/return`: Devolver livro emprestado

## 🔐 Autenticação

A API utiliza JWT para autenticação. Para acessar rotas protegidas:

1. Obtenha um token via endpoint de login
2. Inclua o token no cabeçalho de requisições:

```sh
Authorization: Bearer seu_token_jwt
```

## 📝 Exemplos de Uso

### Registrar um usuário

```sh
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name": "João Silva", "email": "joao@exemplo.com", "password": "senha123"}'
```

### Login

```sh
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "joao@exemplo.com", "password": "senha123"}'
```

### Listar livros disponíveis

```sh
curl http://localhost:8080/api/books
```

### Criar empréstimo (autenticado)

```sh
curl -X POST http://localhost:8080/api/loans \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer SEU_TOKEN_JWT" \
  -d '{"book_id": 1, "return_date": "2025-04-18T00:00:00Z"}'
```

## 🧪 Executando Testes

```sh
# Executar todos os testes
go test ./...

# Executar testes com cobertura
go test -cover ./...

# Gerar relatório de cobertura
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🐳 Comandos Docker

```sh
# Iniciar container PostgreSQL
docker run --name postgres-library -e POSTGRES_PASSWORD=postgres -e POSTGRES_USER=postgres -e POSTGRES_DB=library_api -p 5432:5432 -d postgres

# Verificar se o container está rodando
docker ps
```

## 📄 Licença

Este projeto está sob a licença MIT.

Desenvolvido como projeto de estudos em Go.

