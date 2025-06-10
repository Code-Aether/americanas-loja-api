# americanas-loja-api

> API RESTful para e-commerce inspirado na Americanas, desenvolvida com Go, Gin, GORM e arquitetura limpa.

[![CI/CD Pipeline](https://github.com/Code-Aether/americanas-loja-api/actions/workflows/ci.yml/badge.svg)](https://github.com/Code-Aether/americanas-loja-api/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Code-Aether/americanas-loja-api)](https://goreportcard.com/report/github.com/Code-Aether/americanas-loja-api)
[![codecov](https://codecov.io/gh/Code-Aether/americanas-loja-api/branch/main/graph/badge.svg)](https://codecov.io/gh/Code-Aether/americanas-loja-api)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org)
[![API Documentation](https://img.shields.io/badge/API-Swagger-green)](http://localhost:8080/swagger/index.html)

## Índice
- [Quick Start](#quick-start)
- [Instalação](#️instalação)
- [Configuração](#configuração)
- [Documentação da API](#documentação-da-api)
- [Testes](#testes)
- [Docker](#docker)

## Como Executar a Aplicação

Este guia descreve como configurar e executar a aplicação em seu ambiente de desenvolvimento. Existem dois métodos principais: localmente com Go e SQLite, ou via Docker Compose com PostgreSQL e Redis.

## Instalação

### Pré-requisitos

Antes de começar, certifique-se de que você tem as seguintes ferramentas instaladas:

* **Go:** Versão 1.23 ou superior ([página de download](https://golang.org/dl/))
* **Docker e Docker Compose:** Para a execução em contêineres ([página de download](https://www.docker.com/products/docker-desktop/))
* **Git:** Para clonar o repositório ([página de download](https://git-scm.com/downloads))
* **Make:** Para facilitar a compilação, geração da documentação ([página de download](https://www.gnu.org/software/make/))

## Quick Start

### Gerar JWT secret
```bash
$ ./scripts/generate-secret.sh > jwt_secret.txt
```
### Gerar senha do postgres
```bash
$ echo -n "senha-super-secreta" > postgres_password.txt
```

### Configuração HTTPS para Desenvolvimento Local

Este projeto usa um certificado autoassinado para HTTPS em ambiente de desenvolvimento. Para gerar os arquivos necessários (`cert.pem` e `key.pem`), execute o seguinte comando na raiz do projeto. 

```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem \
  -sha256 -days 365 -nodes \
  -subj "CN=localhost"
```
---

### Instalação Manual

```bash
# 1. Clone o repositório
git clone https://github.com/Code-Aether/americanas-loja-api.git
cd americanas-loja-api

# Copiar .env.example e configurar .env
cp .env.example .env

# 2. Instalar dependências
go mod download
go mod tidy

# 3. Instalar ferramentas de desenvolvimento
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/cosmtrek/air@latest

# 4. Gerar documentação Swagger
make swagger

# 5. Executar testes
make test

# 6. Executar aplicação
make run

# Servidor vai estar rodando em localhost:8080

```

### Usando Docker
```bash
docker compose up -d

#Servidor vai estar rodando em localhost:443
```

## 📚 Documentação da API

### Swagger UI

Acesse a documentação interativa em: 
- Sem docker: **http://localhost:8080/swagger/index.html**
- Com docker: **https://localhost/swagger/index.html**

### Endpoints Principais

#### Autenticação

```bash
# Registro
POST /api/v1/auth/register
{
  "name": "João Silva",
  "email": "joao@teste.com", 
  "password": "123456"
}

# Login
POST /api/v1/auth/login
{
  "email": "joao@teste.com",
  "password": "123456"
}

# Perfil (autenticado)
GET /api/v1/user/profile
Authorization: Bearer 
```

#### 📦 Produtos

```bash
# Listar produtos
GET /api/v1/products?page=1&limit=10&category=Eletrônicos

# Obter produto específico
GET /api/v1/products/1

# Criar produto (autenticado)
POST /api/v1/products
Authorization: Bearer 
{
  "name": "iPhone 15",
  "price": 8999.99,
  "stock": 10,
  "category": "Eletrônicos",
  "sku": "IPHONE-15"
}

# Atualizar produto (autenticado)
PUT /api/v1/products/1
Authorization: Bearer 

# Deletar produto (apenas admin)
DELETE /api/v1/products/1
Authorization: Bearer 
```

### Exemplos de Uso

```bash
# 1. Registrar usuário
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@test.com","password":"123456"}' \
  | jq -r '.data.token')

# 2. Criar produto
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Produto Teste","price":99.99,"stock":10,"sku":"TEST-001","category":"Teste"}'

# 3. Listar produtos
curl http://localhost:8080/api/v1/products
```

## Testes

### Executar Testes

```bash
# Todos os testes
make test

# Testes com coverage
make coverage

# Testes em modo watch
make test-watch

# Testes de race condition
make test-race

# Benchmarks
make benchmark

# Pipeline completo
make test-all
```

### Coverage Report

```bash
# Gerar relatório HTML
make coverage

# Abrir no browser
open coverage.html
```

### Estrutura de Testes

- **Unit Tests** - `*_test.go` em cada pacote
- **Integration Tests** - `internal/integration/`
- **Test Utilities** - `internal/testutils/`
- **Mocks** - Gerados automaticamente

## Docker

### Build Local

```bash
# Build da imagem
make docker-build

# Executar container
make docker-run

# Ver logs
make docker-logs
```

## 🔄 CI/CD

### GitHub Actions

O projeto inclui pipeline completo de CI/CD:

- ✅ **Lint & Format** - golangci-lint, gofmt
- ✅ **Tests** - Unit tests com coverage
- ✅ **Security** - Verificação de vulnerabilidades
- ✅ **Build** - Multi-platform (Linux, Windows, macOS)
- ✅ **Docker** - Build e push automatizado
- ✅ **Release** - Releases automáticos

### Pipeline Stages

```yaml
lint → test → security → build → docker → release
```

## 🚧 Em Desenvolvimento

- **Sistema de Pagamentos** - Integração com gateways
- **Notificações** - Email e push notifications  
- **Analytics** - Métricas e relatórios
- **Busca Avançada** - Elasticsearch
- **Rate Limiting** - Redis
