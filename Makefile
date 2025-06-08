.PHONY: help install run build test clean swagger docker lint install-tool install-dev-tools

help:
	@echo "Comandos disponiveis:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKELIST_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf " \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install:
	@echo "Instalando dependências..."
	go mod download
	go mod tidy
	@echo "Dependências instaladas!"

install-tools:
	@echo "Instalando ferramentas..."
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Ferramentas instaladas!"

install-dev-tools:
	@echo "Instalando ferramentas de desenvolvimento..."
	go install github.com/cosmtrek/air@latest
	@echo "Ferramentas de desenvolvimento instaladas!"

run: swagger
	@echo "Iniciando servidor..."
	go run cmd/server/main.go

dev: air
	@echo "Modo desenvolvimento (usando air)..."
	@if commnad -v air > /dev/null; then \
		air; \
		@echo Hot reload funcionando...
	fi

build: swagger 
	@echo "Compilando projeto..."
	go build -o bin/api cmd/server/main.go
	@echo "Compilado em ./bin/api"

build-linux: swagger
	@echo "Compilando para Linux..."
	GOOS=linux GOARCH=amd64 go build -o bin/api-linux cmd/server/main.go
	@echo "Compilado em ./bin/api-linux"


air:
	@echo "Iniciando hot reload com air..."
	@if ! command -v air > /dev/null; then \
		echo "air não encontrado. Instalando..."; \
		go install github.com/cosmtrek/air@latest \
	fi

swagger:
	@echo "Gerando documentação Swagger..."
	@if ! command -v swag > /dev/null; then \
		echo "swag CLI não encontrado. Instalando..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	swag init -g cmd/server/main.go -o ./docs
	@echo "Documentação gerada!"
	@echo "   • Swagger UI: http://localhost:8080/swagger/index.html"

docs: swagger

test: ## Executa testes
	@echo "Executando testes..."
	go test -v ./...

test-coverage:
	@echo "Executando testes com coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-race:
	@echo "Executando testes (race detection)..."
	go test -race -v ./...

benchmark: ## Executa benchmarks
	@echo "Executando benchmarks..."
	go test -bench=. -benchmem ./...

lint: ## Executa linter
	@echo "Executando linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint não encontrado. Instalando..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

fmt: ## Formata código
	@echo "Formatando código..."
	go fmt ./...
	@echo "Código formatado!"

vet: ## Executa go vet
	@echo "🔍 Executando go vet..."
	go vet ./...

# 🗄️ Banco de Dados
db-reset: ## Reseta banco de dados
	@echo "Resetando banco de dados..."
	rm -f products.db
	@echo "Banco resetado!"

db-migrate: ## Executa migrações (placeholder)
	@echo "Executando migrações..."
	@echo "️ TODO: Implementar migrações"

# 🐳 Docker
docker-build: ## Builda imagem Docker
	@echo "Buildando imagem Docker..."
	docker build -t americanas-loja-api .
	@echo "Imagem construída!"

docker-run: docker-build ## Executa com Docker
	@echo "Executando com Docker..."
	docker run -p 8080:8080 americanas-loja-api

clean:
	@echo "Limpando arquivos..."
	rm -rf bin/
	rm -rf docs/
	rm -f coverage.out coverage.html
	rm -f products.db
	@echo "Limpeza concluída!"

setup: install install-tools swagger
	@echo ""
	@echo "Setup completo! Comandos úteis:"
	@echo "   • make run      - Iniciar servidor"
	@echo "   • make test     - Executar testes"
	@echo "   • make swagger  - Gerar documentação"
	@echo "   • make lint     - Verificar qualidade"
	@echo ""

status:
	@echo "Status do Projeto:"
	@echo "   • Go version: $(shell go version)"
	@echo "   • Módulo: $(shell head -1 go.mod)"
	@echo "   • Arquivos Go: $(shell find . -name '*.go' | wc -l)"
	@echo "   • Dependências: $(shell go list -m all | wc -l)"
	@echo "   • Swagger: $(shell [ -f docs/swagger.json ] && echo '✅ Gerado' || echo '❌ Não gerado')"
	@echo "   • Testes: $(shell go list ./... | wc -l) pacotes"