# 🎬 Setup para Amanhã - Semana 2 Finalização

## 📋 Checklist de Amanhã

### Fase 1: Validação (30 min)
```bash
# Terminal 1
cd /home/caiofaria/vert/packages/helper/golang_vert_helper

# Resolver dependências
go mod tidy

# Tentar compilar
go build ./...

# Se houver erros, corrigir imports
```

### Fase 2: PostgreSQL Setup (30 min)
```bash
# Opção A: Docker (recomendado para testes)
docker run --name vert-helper-test \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=vert_helper_test \
  -p 5432:5432 \
  -d postgres:15-alpine

# Opção B: PostgreSQL local já instalado
# Criar database manualmente
psql -U postgres -c "CREATE DATABASE vert_helper_test;"

# Verificar conexão
psql -h localhost -U postgres -d vert_helper_test -c "SELECT 1;"
```

### Fase 3: Rodar Testes (30 min)
```bash
# Definir env vars (se não estiverem default)
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=postgres
export TEST_DB_PASSWORD=postgres
export TEST_DB_NAME=vert_helper_test

# Rodar testes
go test ./tests/integration/... -v -count=1

# Esperado: 15 testes passando ✅
```

### Fase 4: Testes Incompletos (1 hora)
Adicionar testes para 3 repositories que faltam:
- ServiceHealthRepository (2-3 testes)
- ActionExecutionRepository (2-3 testes)  
- WorkerSnapshotRepository (2-3 testes)

**Arquivo:** `tests/integration/repository_test.go`

**Padrão a seguir:**
```go
func TestServiceHealthRepository_Create(t *testing.T) {
	db := testdb.Setup(t)
	defer db.Cleanup()

	healthRepo := NewServiceHealthRepository(db.DB)
	ctx := context.Background()

	// Create service first
	service := db.CreateService(ctx, "Test Service")

	health := &domain.ServiceHealth{
		ServiceID: service.ID,
		Status:    domain.HealthStatusHealthy,
		Message:   "Test",
	}

	if err := healthRepo.Create(ctx, health); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if health.ID == "" {
		t.Error("Health ID should not be empty")
	}
}
```

---

## 🔍 Possíveis Erros & Soluções

### Erro: `undefined: NewServiceRepository`
**Solução:** Importar `github.com/vert/golang_vert_helper/internal/repository`

### Erro: `pq: role "postgres" does not exist`
**Solução:** Usar usuário correto do PostgreSQL (ex: `postgres`)

### Erro: `failed to connect to test database`
**Solução:** Verificar se PostgreSQL está rodando
```bash
# Docker
docker ps | grep postgres

# Local
pg_isready -h localhost
```

### Erro: `migration state is dirty`
**Solução:** Deletar DB e recriar
```bash
docker rm -f vert-helper-test  # Se Docker
# ou
psql -U postgres -c "DROP DATABASE vert_helper_test;"
```

---

## 📊 Estado Esperado Após Amanhã

```
✅ Semana 1: Setup & Domain - COMPLETO
✅ Semana 2: Database & Repositories - COMPLETO (após validação)
  ✅ Config + GORM
  ✅ 7 Repositories CRUD
  ✅ Migrations SQL
  ✅ TestDB helpers
  ✅ 15+ Integration tests
  ✅ Factory Pattern
  ✅ Compilação verificada
  ✅ Testes passando

⏭️  Semana 3: Core Services (próximo)
  - HealthService
  - ActionService
  - SyncService
  - Built-in health checkers
```

---

## 🎯 Tempo Estimado

| Tarefa | Tempo | Acumulado |
|--------|-------|-----------|
| go mod tidy + build | 5 min | 5 min |
| PostgreSQL setup | 10 min | 15 min |
| Rodar testes | 10 min | 25 min |
| Corrigir falhas | 20 min | 45 min |
| Adicionar 3 repos testes | 45 min | 1:30h |
| Verificação final | 15 min | 1:45h |

**Total: ~2 horas** para finalizar Semana 2 completamente

---

## 📁 Arquivos para revisar antes de amanhã

Se quiser revisar hoje:
1. **SEMANA_2_PROGRESS.md** - Documentação completa
2. **QUICK_SUMMARY.md** - Resumo executivo
3. **internal/repository/repositories.go** - Principais implementações
4. **tests/integration/repository_test.go** - Exemplos de testes

---

## ✨ Bom trabalho!

Semana 2 está 70% pronta, falta apenas validação e completar testes.

Amanhã: Compilação ✅ → Testes ✅ → Semana 3 🚀
