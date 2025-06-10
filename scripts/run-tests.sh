#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_section() {
    echo -e "\n${BLUE}=== $1 ===${NC}\n"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

if [ ! -f "go.mod" ]; then
    print_error "Execute este script na pasta raiz do projeto (onde está o go.mod)"
    exit 1
fi

echo -e "${BLUE}"
echo "🧪 =================================="
echo "   AMERICANAS LOJA API - TESTES"
echo "   $(date)"
echo "==================================="
echo -e "${NC}"

COVERAGE_THRESHOLD=70
VERBOSE=${VERBOSE:-false}
FAST=${FAST:-false}
CI=${CI:-false}

while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --fast|-f)
            FAST=true
            shift
            ;;
        --ci)
            CI=true
            shift
            ;;
        --coverage-threshold)
            COVERAGE_THRESHOLD="$2"
            shift 2
            ;;
        --help|-h)
            echo "Uso: $0 [opções]"
            echo ""
            echo "Opções:"
            echo "  --verbose, -v          Output verbose"
            echo "  --fast, -f             Executa apenas testes rápidos"
            echo "  --ci                   Modo CI (formato especial)"
            echo "  --coverage-threshold   Limiar mínimo de coverage (padrão: 70)"
            echo "  --help, -h             Mostra esta ajuda"
            echo ""
            echo "Variáveis de ambiente:"
            echo "  VERBOSE=true           Mesmo que --verbose"
            echo "  FAST=true              Mesmo que --fast"
            echo "  CI=true                Mesmo que --ci"
            exit 0
            ;;
        *)
            print_error "Opção desconhecida: $1"
            exit 1
            ;;
    esac
done

TEST_FLAGS="-v"
if [ "$VERBOSE" = "false" ] && [ "$CI" = "true" ]; then
    TEST_FLAGS=""
fi

if [ "$FAST" = "true" ]; then
    TEST_FLAGS="$TEST_FLAGS -short"
fi

run_with_output() {
    local cmd="$1"
    local description="$2"
    
    if [ "$VERBOSE" = "true" ] || [ "$CI" = "false" ]; then
        echo "💻 Executando: $cmd"
        eval $cmd
    else
        eval $cmd > /tmp/test_output 2>&1
        if [ $? -eq 0 ]; then
            print_success "$description"
        else
            print_error "$description"
            cat /tmp/test_output
            return 1
        fi
    fi
}

print_section "VERIFICAÇÕES INICIAIS"

echo "Verificando ambiente..."
go version
echo "Verificando dependências..."
go mod verify
print_success "Ambiente OK"

if [ "$FAST" = "false" ]; then
    print_section "LINTING E FORMATAÇÃO"
    
    echo "Verificando formatação..."
    UNFORMATTED=$(gofmt -l .)
    if [ -n "$UNFORMATTED" ]; then
        print_error "Arquivos não formatados encontrados:"
        echo "$UNFORMATTED"
        exit 1
    fi
    print_success "Formatação OK"
    
    echo "Executando go vet..."
    if ! go vet ./...; then
        print_error "go vet falhou"
        exit 1
    fi
    print_success "go vet OK"
    
    # golangci-lint se disponível
    if command -v golangci-lint &> /dev/null; then
        echo "Executando golangci-lint..."
        if ! golangci-lint run --timeout=3m; then
            print_error "golangci-lint falhou"
            exit 1
        fi
        print_success "golangci-lint OK"
    else
        print_warning "golangci-lint não instalado (opcional)"
    fi
fi

# 3. Compilação
print_section "COMPILAÇÃO"

echo "Verificando se compila..."
if ! go build -o /tmp/api-test cmd/server/main.go; then
    print_error "Falha na compilação"
    exit 1
fi
rm -f /tmp/api-test
print_success "Compilação OK"

# 4. Testes unitários
print_section "TESTES UNITÁRIOS"

echo "🧪 Executando testes unitários..."
TEST_PACKAGES=$(go list ./... | grep -v /vendor/)

if [ "$CI" = "true" ]; then
    # Modo CI - com coverage e race detection
    if ! go test -race -coverprofile=coverage.out -covermode=atomic $TEST_FLAGS $TEST_PACKAGES; then
        print_error "Testes unitários falharam"
        exit 1
    fi
else
    # Modo desenvolvimento
    if ! go test $TEST_FLAGS $TEST_PACKAGES; then
        print_error "Testes unitários falharam"
        exit 1
    fi
fi

print_success "Testes unitários OK"

# 5. Análise de coverage
if [ -f "coverage.out" ]; then
    print_section "ANÁLISE DE COVERAGE"
    
    echo "Calculando coverage..."
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    
    echo "Coverage total: ${COVERAGE}%"
    
    if [ "$CI" = "true" ]; then
        echo "::set-output name=coverage::${COVERAGE}"
    fi
    
    # Verificar threshold
    if (( $(echo "$COVERAGE < $COVERAGE_THRESHOLD" | bc -l) )); then
        print_error "Coverage muito baixo: ${COVERAGE}% (mínimo: ${COVERAGE_THRESHOLD}%)"
        exit 1
    fi
    
    print_success "Coverage OK: ${COVERAGE}%"
    
    # Gerar relatório HTML se não for CI
    if [ "$CI" = "false" ]; then
        go tool cover -html=coverage.out -o coverage.html
        echo "📄 Relatório HTML gerado: coverage.html"
    fi
fi

# 6. Race condition detection (se não for modo fast)
if [ "$FAST" = "false" ] && [ "$CI" = "false" ]; then
    print_section "DETECÇÃO DE RACE CONDITIONS"
    
    echo "Executando testes com race detection..."
    if ! go test -race $TEST_PACKAGES; then
        print_error "Race conditions detectadas"
        exit 1
    fi
    print_success "Nenhuma race condition detectada"
fi

# 7. Benchmarks (se não for modo fast)
if [ "$FAST" = "false" ] && [ "$CI" = "false" ]; then
    print_section "BENCHMARKS"
    
    echo "Executando benchmarks..."
    BENCH_OUTPUT=$(go test -bench=. -benchmem $TEST_PACKAGES 2>/dev/null | grep -E "^Benchmark")
    
    if [ -n "$BENCH_OUTPUT" ]; then
        echo "$BENCH_OUTPUT"
        print_success "Benchmarks executados"
    else
        print_warning "Nenhum benchmark encontrado"
    fi
fi

# 8. Verificação de vulnerabilidades (se disponível)
if command -v govulncheck &> /dev/null && [ "$FAST" = "false" ]; then
    print_section "VERIFICAÇÃO DE VULNERABILIDADES"
    
    echo "🔒 Verificando vulnerabilidades..."
    if ! govulncheck ./...; then
        print_error "Vulnerabilidades encontradas"
        exit 1
    fi
    print_success "Nenhuma vulnerabilidade encontrada"
fi

# 9. Análise de dependências
if [ "$FAST" = "false" ]; then
    print_section "ANÁLISE DE DEPENDÊNCIAS"
    
    echo "📦 Verificando dependências não utilizadas..."
    if command -v go-mod-outdated &> /dev/null; then
        go list -u -m all | go-mod-outdated -update -direct
    fi
    
    echo "🧹 Verificando go.mod..."
    go mod tidy
    if [ -n "$(git status --porcelain go.mod go.sum)" ]; then
        print_warning "go.mod/go.sum precisa ser atualizado (execute 'go mod tidy')"
    else
        print_success "go.mod OK"
    fi
fi

# 10. Resumo final
print_section "RESUMO FINAL"

echo "Estatísticas dos testes:"
if [ -f "coverage.out" ]; then
    echo "   • Coverage: ${COVERAGE}%"
fi

TOTAL_TESTS=$(grep -E "^=== RUN" /tmp/test_output 2>/dev/null | wc -l || echo "N/A")
echo "   • Testes executados: $TOTAL_TESTS"

echo "   • Pacotes testados: $(echo "$TEST_PACKAGES" | wc -l)"

# Calcular tempo total
if [ -n "$START_TIME" ]; then
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    echo "   • Tempo total: ${DURATION}s"
fi

echo ""
print_success "TODOS OS TESTES PASSARAM! 🎉"

echo ""
echo -e "${GREEN}Finalizado com sucesso.${NC}"