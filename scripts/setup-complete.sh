#!/bin/bash

set -e 

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' 

print_header() {
    echo -e "\n${PURPLE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${PURPLE}║                                                              ║${NC}"
    echo -e "${PURPLE}║  AMERICANAS LOJA API - SETUP COMPLETO E VERIFICAÇÃO          ║${NC}"
    echo -e "${PURPLE}║                                                              ║${NC}"
    echo -e "${PURPLE}╚══════════════════════════════════════════════════════════════╝${NC}"
}

print_section() {
    echo -e "\n${BLUE}═══ $1 ═══${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}! $1${NC}"
}

print_error() {
    echo -e "${RED}x $1${NC}"
}

print_info() {
    echo -e "${CYAN}i $1${NC}"
}

if [ ! -f "go.mod" ]; then
    print_error "Execute este script na pasta raiz do projeto (onde está o go.mod)"
    exit 1
fi

# Banner
print_header

echo -e "${CYAN}Iniciado em: $(date)${NC}"
echo -e "${CYAN}Diretório: $(pwd)${NC}"
echo -e "${CYAN}Usuário: $(whoami)${NC}"
echo -e "${CYAN}Sistema: $(uname -s)${NC}"

# Variáveis
PROJECT_NAME="americanas-loja-api"
GO_VERSION="1.23"
REQUIRED_TOOLS=("go" "git" "make")

# 1. Verificações iniciais
print_section "VERIFICAÇÕES INICIAIS"

echo "Verificando pré-requisitos..."

# Verificar Go
if command -v go &> /dev/null; then
    GO_CURRENT=$(go version | grep -oE "go[0-9]+\.[0-9]+" | head -1)
    print_success "Go encontrado: $(go version)"
    
    # Verificar versão mínima
    if [[ "$(printf '%s\n' "$GO_VERSION" "$GO_CURRENT" | sort -V | head -n1)" = "$GO_VERSION" ]]; then
        print_success "Versão do Go OK (>= $GO_VERSION)"
    else
        print_warning "Versão do Go antiga. Recomendado: >= $GO_VERSION"
    fi
else
    print_error "Go não encontrado. Instale: https://golang.org/dl/"
    exit 1
fi

# Verificar outras ferramentas
for tool in "${REQUIRED_TOOLS[@]}"; do
    if command -v "$tool" &> /dev/null; then
        print_success "$tool encontrado"
    else
        print_error "$tool não encontrado"
        exit 1
    fi
done

# Verificar se é um projeto Go válido
if [ -f "go.mod" ]; then
    print_success "Projeto Go válido encontrado"
    echo "   • Módulo: $(head -1 go.mod)"
else
    print_error "go.mod não encontrado"
    exit 1
fi

# 2. Instalação de dependências
print_section "INSTALAÇÃO DE DEPENDÊNCIAS"

echo "Baixando dependências do Go..."
go mod download
go mod tidy
print_success "Dependências do Go instaladas"

echo "Instalando ferramentas de desenvolvimento..."

# Lista de ferramentas
declare -A tools=(
    ["swag"]="github.com/swaggo/swag/cmd/swag@latest"
    ["golangci-lint"]="github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
)

for tool in "${!tools[@]}"; do
    echo "   • Instalando $tool..."
    if go install "${tools[$tool]}"; then
        print_success "$tool instalado"
    else
        print_warning "Falha ao instalar $tool (opcional)"
    fi
done

# 3. Geração de documentação
print_section "GERAÇÃO DE DOCUMENTAÇÃO"

echo "Gerando documentação Swagger..."
if command -v swag &> /dev/null; then
    if swag init -g cmd/server/main.go -o ./docs --parseInternal --parseDepth 2; then
        print_success "Swagger gerado com sucesso"
        if [ -f "docs/swagger.json" ]; then
            ENDPOINTS=$(grep -c '^\s*"\/[^"]*": {' docs/swagger.json)
            echo "   • Endpoints documentados: $ENDPOINTS"
        fi
    else
        print_error "Falha ao gerar Swagger"
        exit 1
    fi
else
    print_warning "swag não instalado, pulando geração de documentação"
fi

# 4. Compilação
print_section "COMPILAÇÃO"

echo "Compilando projeto..."
if go build -o bin/api cmd/server/main.go; then
    print_success "Compilação bem-sucedida"
    
    # Verificar tamanho do binário
    if [ -f "bin/api" ]; then
        SIZE=$(du -h bin/api | cut -f1)
        echo "   • Tamanho do binário: $SIZE"
    fi
else
    print_error "Falha na compilação"
    exit 1
fi

# 5. Execução de testes
print_section "EXECUÇÃO DE TESTES"

echo "Executando testes..."
if go test -v ./...; then
    print_success "Todos os testes passaram"
    
    # Executar testes com coverage
    echo "Calculando coverage..."
    if go test -coverprofile=coverage.out ./...; then
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        print_success "Coverage: $COVERAGE"
        
        # Gerar relatório HTML
        go tool cover -html=coverage.out -o coverage.html
        echo "   • Relatório HTML: coverage.html"
    else
        print_warning "Falha ao calcular coverage"
    fi
else
    print_error "Alguns testes falharam"
    exit 1
fi

# 6. Verificação de qualidade
print_section "VERIFICAÇÃO DE QUALIDADE"

# Formatação
echo "Verificando formatação..."
UNFORMATTED=$(gofmt -l . | grep -v vendor || true)
if [ -z "$UNFORMATTED" ]; then
    print_success "Código bem formatado"
else
    print_warning "Alguns arquivos precisam de formatação:"
    echo "$UNFORMATTED"
fi

# Vet
echo "Executando go vet..."
if go vet ./...; then
    print_success "go vet OK"
else
    print_warning "go vet encontrou problemas"
fi

# Linter (se disponível)
if command -v golangci-lint &> /dev/null; then
    echo "Executando golangci-lint..."
    if golangci-lint run --timeout=3m; then
        print_success "golangci-lint OK"
    else
        print_warning "golangci-lint encontrou problemas"
    fi
else
    print_warning "golangci-lint não instalado (recomendado)"
fi

# 7. Teste de inicialização
print_section "TESTE DE INICIALIZAÇÃO"

echo "Testando inicialização do servidor..."

# Iniciar servidor em background
export PORT=8081  # Usar porta diferente para evitar conflitos
export JWT_SECRET=$(cat jwt_secret.txt)
export GIN_MODE="release"

./bin/api &
SERVER_PID=$!
echo "   • Servidor iniciado (PID: $SERVER_PID)"

# Aguardar servidor inicializar
sleep 3

# Testar health check
echo "Testando health check..."
if curl -s -f http://localhost:8081/health > /dev/null; then
    print_success "Health check OK"
    
    # Testar Swagger
    echo " Testando Swagger UI..."
    if curl -s -f http://localhost:8081/swagger/index.html > /dev/null; then
        print_success "Swagger UI OK"
    else
        print_warning "Swagger UI não acessível"
    fi
    
else
    print_error "Health check falhou"
fi

# Parar servidor
echo "Parando servidor de teste..."
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true
print_success "Servidor parado"

# 8. Verificação de arquivos essenciais
print_section "VERIFICAÇÃO DE ARQUIVOS"

REQUIRED_FILES=(
    "go.mod"
    "go.sum"
    "Makefile"
    "Dockerfile"
    "README.md"
    "cmd/server/main.go"
    "internal/handlers"
    "internal/services"
    "internal/repository"
    "docs/swagger.json"
    ".github/workflows"
)

echo "Verificando arquivos essenciais..."
for file in "${REQUIRED_FILES[@]}"; do
    if [ -e "$file" ]; then
        print_success "$file ✓"
    else
        print_warning "$file não encontrado"
    fi
done

# 9. Relatório final
print_section "RELATÓRIO FINAL"

echo "Estatísticas do projeto:"
echo "   • Arquivos Go: $(find . -name '*.go' -not -path './vendor/*' | wc -l)"
echo "   • Linhas de código: $(find . -name '*.go' -not -path './vendor/*' -exec wc -l {} + | tail -1 | awk '{print $1}')"
echo "   • Arquivos de teste: $(find . -name '*_test.go' | wc -l)"
echo "   • Dependências: $(go list -m all | wc -l)"
echo "   • Tamanho do projeto: $(du -sh . | cut -f1)"

if [ -f "coverage.out" ]; then
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    echo "   • Coverage de testes: $COVERAGE"
fi

if [ -f "docs/swagger.json" ]; then
    ENDPOINTS=$(grep -c '^\s*"\/[^"]*": {' docs/swagger.json)
    echo "   • Endpoints da API: $ENDPOINTS"
fi

echo ""
print_success "SETUP COMPLETO!"

exit 0