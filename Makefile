# Makefile para Americanas Loja API - Versão Completa com Testes

.PHONY: help install run build test clean swagger docker lint coverage

# 🎯 Default target
help: ## Mostra esta ajuda
	@echo "🚀 Americanas Loja API - Comandos Disponíveis:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# 📦 Instalação e Setup
install: ## Instala dependências
	@echo "📦 Instalando dependências..."
	go mod download
	go mod tidy
	@echo "✅ Dependências instaladas!"

install-tools: ## Instala ferramentas de desenvolvimento
	@echo "🛠️ Instalando ferramentas..."
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest
	@echo "✅ Ferramentas instaladas!"

setup: install install-tools swagger ## Setup completo do projeto
	@echo ""
	@echo "🎉 Setup completo! Comandos úteis:"
	@echo "   • make run         - Iniciar servidor"
	@echo "   • make test        - Executar testes"
	@echo "   • make test-watch  - Testes com watch"
	@echo "   • make coverage    - Coverage report"
	@echo "   • make swagger     - Gerar documentação"
	@echo "   • make lint        - Verificar qualidade"
	@echo ""

# 🚀 Execução
run: swagger ## Executa o servidor
	@echo "🚀 Iniciando servidor..."
	go run cmd/server/main.go

dev: ## Executa em modo desenvolvimento com hot reload
	@echo "🔄 Modo desenvolvimento (hot reload)..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "💡 Instale air para hot reload: make install-tools"; \
		make run; \
	fi

# 🏗️ Build
build: swagger ## Compila o projeto
	@echo "🏗️ Compilando projeto..."
	CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/api cmd/server/main.go
	@echo "✅ Compilado em ./bin/api"

build-linux: swagger ## Compila para Linux
	@echo "🐧 Compilando para Linux..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/api-linux cmd/server/main.go
	@echo "✅ Compilado em ./bin/api-linux"

build-windows: swagger ## Compila para Windows
	@echo "🪟 Compilando para Windows..."
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/api.exe cmd/server/main.go
	@echo "✅ Compilado em ./bin/api.exe"

build-all: build build-linux build-windows ## Compila para todas as plataformas
	@echo "🌍 Builds para todas as plataformas concluídos!"
	@ls -la bin/

# 📚 Documentação
swagger: ## Gera documentação Swagger
	@echo "📚 Gerando documentação Swagger..."
	@if ! command -v swag > /dev/null; then \
		echo "❌ swag CLI não encontrado. Instalando..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	swag init -g cmd/server/main.go -o ./docs --parseInternal --parseDepth 2
	@echo "✅ Documentação gerada!"
	@echo "   • Swagger UI: http://localhost:8080/swagger/index.html"

docs: swagger ## Alias para swagger

docs-serve: swagger ## Serve documentação localmente
	@echo "📖 Servindo documentação..."
	@echo "   • Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "   • Execute 'make run' em outro terminal"

# 🧪 Testes
test: ## Executa todos os testes
	@echo "🧪 Executando testes..."
	go test -v ./...

test-short: ## Executa testes rápidos (sem integração)
	@echo "⚡ Executando testes rápidos..."
	go test -short -v ./...

test-watch: ## Executa testes em modo watch
	@echo "👀 Testes em modo watch (Ctrl+C para parar)..."
	@if command -v air > /dev/null; then \
		air -c .air-test.toml; \
	else \
		echo "💡 Instale air para watch mode: make install-tools"; \
		echo "🔄 Executando testes em loop manual..."; \
		while true; do \
			clear; \
			echo "🧪 Executando testes - $(date)"; \
			go test -v ./...; \
			echo ""; \
			echo "⏳ Aguardando alterações... (Ctrl+C para parar)"; \
			sleep 5; \
		done; \
	fi

test-coverage: ## Executa testes com coverage completo
	@echo "🧪 Executando testes com coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report gerado!"
	@echo "   • Terminal: coverage.out"
	@echo "   • HTML: coverage.html"
	@echo "   • Resumo acima ☝️"

coverage: test-coverage ## Alias para test-coverage

test-race: ## Executa testes verificando race conditions
	@echo "🏁 Executando testes com race detection..."
	go test -race -v ./...

test-parallel: ## Executa testes em paralelo
	@echo "⚡ Executando testes em paralelo..."
	go test -parallel 4 -v ./...

benchmark: ## Executa benchmarks
	@echo "⚡ Executando benchmarks..."
	go test -bench=. -benchmem ./...

test-all: test test-race benchmark ## Executa todos os tipos de teste
	@echo "🎯 Todos os testes concluídos!"

# 🔍 Qualidade de Código
lint: ## Executa linter completo
	@echo "🔍 Executando linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "❌ golangci-lint não encontrado. Instalando..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run --timeout=5m; \
	fi

lint-fix: ## Executa linter e corrige problemas automaticamente
	@echo "🔧 Executando linter com correções..."
	golangci-lint run --fix --timeout=5m

fmt: ## Formata código
	@echo "✨ Formatando código..."
	go fmt ./...
	@echo "✅ Código formatado!"

vet: ## Executa go vet
	@echo "🔍 Executando go vet..."
	go vet ./...

tidy: ## Organiza dependências
	@echo "🧹 Organizando dependências..."
	go mod tidy

quality: fmt vet lint ## Executa todas as verificações de qualidade
	@echo "💎 Verificações de qualidade concluídas!"

# 🗄️ Banco de Dados
db-reset: ## Reseta banco de dados
	@echo "🗄️ Resetando banco de dados..."
	rm -f products.db
	@echo "✅ Banco resetado!"

db-seed: ## Popula banco com dados de exemplo
	@echo "🌱 Populando banco com dados de exemplo..."
	go run scripts/seed.go
	@echo "✅ Dados de exemplo inseridos!"

db-backup: ## Faz backup do banco
	@echo "💾 Fazendo backup do banco..."
	cp products.db "backup_$(shell date +%Y%m%d_%H%M%S).db"
	@echo "✅ Backup criado!"

# 🐳 Docker
docker-build: ## Builda imagem Docker
	@echo "🐳 Buildando imagem Docker..."
	docker build -t americanas-loja-api:latest .
	@echo "✅ Imagem construída: americanas-loja-api:latest"

docker-run: docker-build ## Executa com Docker
	@echo "🐳 Executando com Docker..."
	docker run -p 8080:8080 --name americanas-api americanas-loja-api:latest

docker-stop: ## Para container Docker
	@echo "🛑 Parando container..."
	docker stop americanas-api || true
	docker rm americanas-api || true

docker-logs: ## Mostra logs do container
	@echo "📋 Logs do container:"
	docker logs americanas-api

docker-shell: ## Acessa shell do container
	@echo "🐚 Acessando shell do container..."
	docker exec -it americanas-api /bin/sh

# 🧹 Limpeza
clean: ## Limpa arquivos gerados
	@echo "🧹 Limpando arquivos..."
	rm -rf bin/
	rm -rf docs/
	rm -f coverage.out coverage.html
	rm -f *.prof
	rm -f products.db
	rm -f backup_*.db
	@echo "✅ Limpeza concluída!"

clean-docker: ## Limpa imagens Docker
	@echo "🐳 Limpando imagens Docker..."
	docker rmi americanas-loja-api:latest || true
	@echo "✅ Imagens Docker removidas!"

clean-all: clean clean-docker ## Limpeza completa
	@echo "🧹 Limpeza completa concluída!"

# 🚀 CI/CD e Deploy
ci-install: ## Instala dependências para CI
	@echo "🤖 Instalando dependências para CI..."
	go mod download
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

ci-test: ci-install ## Executa testes para CI
	@echo "🤖 Executando testes para CI..."
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

ci-build: ci-install swagger ## Build para CI
	@echo "🤖 Build para CI..."
	go build -ldflags="-s -w" -o bin/api cmd/server/main.go

ci-lint: ## Lint para CI
	@echo "🤖 Lint para CI..."
	golangci-lint run --timeout=5m --out-format=github-actions

ci-security: ## Verificação de segurança para CI
	@echo "🔒 Verificação de segurança..."
	go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth

ci-pipeline: ci-install ci-lint ci-test ci-build ## Pipeline completo de CI
	@echo "🎯 Pipeline de CI concluído!"

# 📊 Relatórios e Métricas
report: ## Gera relatório completo do projeto
	@echo "📊 Gerando relatório do projeto..."
	@echo "# Relatório do Projeto - $(shell date)" > report.md
	@echo "" >> report.md
	@echo "## Estrutura do Projeto" >> report.md
	@find . -name "*.go" -not -path "./vendor/*" | head -20 >> report.md
	@echo "" >> report.md
	@echo "## Estatísticas" >> report.md
	@echo "- Arquivos Go: $(shell find . -name '*.go' | wc -l)" >> report.md
	@echo "- Linhas de código: $(shell find . -name '*.go' -exec wc -l {} + | tail -1)" >> report.md
	@echo "- Dependências: $(shell go list -m all | wc -l)" >> report.md
	@echo "✅ Relatório gerado: report.md"

metrics: ## Mostra métricas do projeto
	@echo "📈 Métricas do Projeto:"
	@echo "   • Arquivos Go: $(shell find . -name '*.go' -not -path "./vendor/*" | wc -l)"
	@echo "   • Linhas de código: $(shell find . -name '*.go' -not -path "./vendor/*" -exec wc -l {} + | tail -1 | awk '{print $$1}')"
	@echo "   • Testes: $(shell find . -name '*_test.go' | wc -l)"
	@echo "   • Dependências: $(shell go list -m all | wc -l)"
	@echo "   • Pacotes: $(shell go list ./... | wc -l)"

status: metrics ## Mostra status completo do projeto
	@echo ""
	@echo "📊 Status do Projeto:"
	@echo "   • Go version: $(shell go version | awk '{print $$3}')"
	@echo "   • Módulo: $(shell head -1 go.mod | awk '{print $$2}')"
	@echo "   • Git branch: $(shell git branch --show-current 2>/dev/null || echo 'N/A')"
	@echo "   • Git status: $(shell git status --porcelain | wc -l) arquivos modificados"
	@echo "   • Swagger: $(shell [ -f docs/swagger.json ] && echo '✅ Gerado' || echo '❌ Não gerado')"
	@echo "   • Última build: $(shell [ -f bin/api ] && stat -c %y bin/api || echo 'Nunca')"

# 🔄 Comandos de Desenvolvimento
dev-setup: setup db-reset ## Setup completo para desenvolvimento
	@echo "🔧 Setup de desenvolvimento concluído!"
	@echo "   • Execute 'make dev' para iniciar com hot reload"
	@echo "   • Execute 'make test-watch' para testes contínuos"

dev-reset: clean setup db-reset ## Reset completo do ambiente
	@echo "🔄 Reset completo do ambiente de desenvolvimento!"

dev-full: dev-setup test swagger ## Setup + testes + documentação
	@echo "🚀 Ambiente de desenvolvimento completamente configurado!"

# 🎯 Comandos Rápidos
quick-test: ## Testes rápidos (apenas unit tests)
	@go test -short -count=1 ./...

quick-build: ## Build rápido sem swagger
	@go build -o bin/api cmd/server/main.go

quick-run: quick-build ## Build e run rápido
	@./bin/api

# 📋 Release
pre-release: clean ci-pipeline ## Preparação para release
	@echo "🏷️ Preparação para release concluída!"
	@echo "   • Todos os testes passaram"
	@echo "   • Linter aprovado"
	@echo "   • Build gerado"
	@echo "   • Pronto para release!"

release: pre-release build-all ## Cria release completo
	@echo "🎉 Release criado!"
	@echo "   • Binários em ./bin/"
	@echo "   • Documentação em ./docs/"
	@echo "   • Coverage em ./coverage.html"
