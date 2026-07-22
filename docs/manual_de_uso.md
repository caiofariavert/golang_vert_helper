# Manual de Uso — Vert Helper Go

Biblioteca Go para monitoramento de saúde de serviços e execução de ações interativas. Integra-se a qualquer aplicação Gin/GORM existente sem impor estrutura ou configuração obrigatória.

---

## Índice

1. [Instalação](#1-instalação)
2. [Inicialização](#2-inicialização)
3. [Registrando Serviços e Health Checks](#3-registrando-serviços-e-health-checks)
4. [Registrando Actions](#4-registrando-actions)
5. [Sincronizando Definições com o Banco](#5-sincronizando-definições-com-o-banco)
6. [Registrando Rotas no Gin](#6-registrando-rotas-no-gin)
7. [Scheduler — Health Checks Periódicos](#7-scheduler--health-checks-periódicos)
8. [Monitoramento de Workers](#8-monitoramento-de-workers)
9. [Callbacks](#9-callbacks)
10. [Referência da API REST](#10-referência-da-api-rest)
11. [Referência da API Go](#11-referência-da-api-go)

---

## 1. Instalação

```bash
go get github.com/caiofariavert/golang_vert_helper
```

**Pré-requisitos:**
- Go 1.21+
- PostgreSQL (ou outro banco suportado pelo GORM)
- Gin já configurado na aplicação

---

## 2. Inicialização

A biblioteca recebe a conexão GORM que **você já possui** — não cria uma nova.

```go
import (
    "github.com/caiofariavert/golang_vert_helper/pkg/helper"
)

// db é sua conexão *gorm.DB existente
h := helper.New(db)
```

### Opções de configuração

```go
import (
    "log/slog"
    "github.com/caiofariavert/golang_vert_helper/pkg/helper"
    "github.com/caiofariavert/golang_vert_helper/internal/domain"
)

h := helper.New(db,
    helper.WithLogger(slog.Default()),

    helper.WithOnHealthCheckFailure(func(ctx context.Context, svc *domain.Service, result *domain.HealthCheckResult) error {
        log.Printf("ALERTA: serviço %s está %s", svc.Name, result.Status)
        return nil
    }),

    helper.WithOnActionExecution(func(ctx context.Context, exec *domain.ActionExecution, result *domain.ActionResult) error {
        log.Printf("Action executada: status=%s", exec.Status)
        return nil
    }),
)
```

---

## 3. Registrando Serviços e Health Checks

Um **serviço** é qualquer dependência que você quer monitorar. Você implementa a interface `HealthChecker` ou usa os checkers prontos.

### Checkers prontos

```go
import healthchecks "github.com/caiofariavert/golang_vert_helper/pkg/health_checks"

// Verifica a conexão com o PostgreSQL via GORM
h.RegisterService("postgres", healthchecks.NewGormPostgresChecker(db))

// Verifica a conexão com S3 (AWS)
h.RegisterService("s3", healthchecks.NewS3Checker(healthchecks.S3Config{
    Region: "sa-east-1",
}))

// Verifica a conexão com MinIO ou LocalStack
h.RegisterService("s3", healthchecks.NewS3Checker(healthchecks.S3Config{
    Region:      "us-east-1",
    IsLocal:     true,
    EndpointURL: "http://localhost:9000",
}))

// Com credenciais estáticas (opcional — se omitidas, o SDK usa variáveis de ambiente,
// IAM role ou ~/.aws/credentials automaticamente)
h.RegisterService("s3", healthchecks.NewS3Checker(healthchecks.S3Config{
    Region:          "sa-east-1",
    AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
    SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
}))
```

### Checker customizado

Implemente a interface `domain.HealthChecker`:

```go
type RedisChecker struct {
    client *redis.Client
}

func (c *RedisChecker) Check(ctx context.Context) (*domain.HealthCheckResult, error) {
    if err := c.client.Ping(ctx).Err(); err != nil {
        return &domain.HealthCheckResult{
            Status:    domain.HealthStatusUnhealthy,
            Message:   err.Error(),
            Timestamp: time.Now(),
        }, nil
    }
    return &domain.HealthCheckResult{
        Status:    domain.HealthStatusHealthy,
        Message:   "Redis OK",
        Timestamp: time.Now(),
    }, nil
}

// Registrar
h.RegisterService("redis", &RedisChecker{client: redisClient})
```

### Status disponíveis

| Constante | Valor | Significado |
|-----------|-------|-------------|
| `domain.HealthStatusHealthy` | `"healthy"` | Serviço funcionando normalmente |
| `domain.HealthStatusDegraded` | `"degraded"` | Funcionando com degradação |
| `domain.HealthStatusUnhealthy` | `"unhealthy"` | Serviço com falha |
| `domain.HealthStatusUnknown` | `"unknown"` | Estado desconhecido |

---

## 4. Registrando Actions

Uma **action** representa uma operação interativa — com formulário de entrada — que pode ser executada pelo usuário. O handler recebe as respostas do formulário e retorna o resultado.

```go
h.RegisterAction("reiniciar-servico", func(
    ctx context.Context,
    action *domain.Action,
    input map[string]interface{},
) (*domain.ActionResult, error) {
    servico := input["servico"].(string)
    motivo  := input["motivo"].(string)

    // Sua lógica aqui
    log.Printf("Reiniciando %s — motivo: %s", servico, motivo)

    return &domain.ActionResult{
        Success: true,
        Message: "Serviço reiniciado com sucesso",
        Data:    map[string]interface{}{"servico": servico},
    }, nil
})
```

O `input` é um `map[string]interface{}` com as respostas do usuário, onde a chave é o `slug` de cada questão definida na action.

---

## 5. Sincronizando Definições com o Banco

Use `Sync` para garantir que as definições de serviços e actions no seu código estejam persistidas no banco. Normalmente chamado na inicialização da aplicação.

```go
import "github.com/caiofariavert/golang_vert_helper/internal/services"

err := h.Sync(ctx, []services.ServiceDefinition{
    {
        Name:        "meu-sistema",
        Description: "Sistema principal",
        Actions: []services.ActionDefinition{
            {
                Slug:        "reiniciar-servico",
                Title:       "Reiniciar Serviço",
                Description: "Reinicia um serviço específico",
                Questions: []services.QuestionDefinition{
                    {
                        Slug:      "servico",
                        InputType: domain.QuestionInputTypeSelect,
                        Label:     "Qual serviço?",
                        Required:  true,
                        Options:   []string{"api", "worker", "scheduler"},
                        Order:     1,
                    },
                    {
                        Slug:      "motivo",
                        InputType: domain.QuestionInputTypeTextarea,
                        Label:     "Motivo do reinício",
                        Required:  true,
                        Order:     2,
                    },
                },
            },
        },
    },
})
```

### Questões condicionais (parent-child)

```go
// A questão "ambiente-custom" só aparece se "ambiente" == "outro"
{
    Slug:        "ambiente",
    InputType:   domain.QuestionInputTypeSelect,
    Label:       "Ambiente",
    Required:    true,
    Options:     []string{"producao", "staging", "outro"},
    Order:       1,
    Children: []services.QuestionDefinition{
        {
            Slug:        "ambiente-custom",
            InputType:   domain.QuestionInputTypeText,
            Label:       "Especifique o ambiente",
            Required:    true,
            ParentSlug:  "ambiente",
            ParentValue: "outro",
            Order:       1,
        },
    },
},
```

---

## 6. Registrando Rotas no Gin

```go
router := gin.Default()

// Sem middleware
h.RegisterRoutes(router, db, nil)

// Com middleware (ex: autenticação JWT)
authMiddleware := func(c *gin.Context) {
    token := c.GetHeader("Authorization")
    if token == "" {
        c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
        return
    }
    c.Next()
}
h.RegisterRoutes(router, db, &authMiddleware)

router.Run(":8080")
```

Todas as rotas são registradas sob o prefixo `/api/helper/v1/`.

---

## 7. Scheduler — Health Checks Periódicos

O scheduler executa health checks automaticamente em background usando cron.

```go
// Configuração padrão: a cada 10 minutos
scheduler := helper.NewScheduler(h, helper.DefaultSchedulerConfig())
defer scheduler.Stop()
```

```go
// Configuração customizada
scheduler := helper.NewScheduler(h, helper.SchedulerConfig{
    HealthCheckCron: "*/5 * * * *", // a cada 5 minutos
})
defer scheduler.Stop()
```

### Expressões cron úteis

| Expressão | Frequência |
|-----------|------------|
| `"*/5 * * * *"` | A cada 5 minutos |
| `"*/10 * * * *"` | A cada 10 minutos (padrão) |
| `"0 * * * *"` | A cada hora |
| `"0 9 * * *"` | Todo dia às 9h |

---

## 8. Monitoramento de Workers

O `WorkerPool` permite registrar goroutines/jobs e monitorar seu estado em tempo real.

### Setup

```go
import healthchecks "github.com/caiofariavert/golang_vert_helper/pkg/health_checks"

pool := healthchecks.NewWorkerPool()

h := helper.New(db,
    helper.WithWorkerPool(pool), // expõe automaticamente em /api/helper/v1/workers/
)
```

### Usando no seu worker

```go
func StartKafkaConsumer(pool *healthchecks.WorkerPool) {
    pool.Register("kafka-consumer-1", "Kafka Consumer Principal")

    go func() {
        for msg := range messages {
            pool.UpdateStatus("kafka-consumer-1", healthchecks.WorkerProcessing, "")

            if err := process(msg); err != nil {
                pool.UpdateStatus("kafka-consumer-1", healthchecks.WorkerFailed, err.Error())
                continue
            }

            pool.UpdateStatus("kafka-consumer-1", healthchecks.WorkerIdle, "")
        }
    }()
}
```

### Status de worker disponíveis

| Constante | Valor | Quando usar |
|-----------|-------|-------------|
| `WorkerRunning` | `"running"` | Worker iniciado, aguardando mensagens |
| `WorkerIdle` | `"idle"` | Aguardando trabalho |
| `WorkerProcessing` | `"processing"` | Processando uma mensagem |
| `WorkerFailed` | `"failed"` | Erro na última execução |
| `WorkerPaused` | `"paused"` | Pausado intencionalmente |
| `WorkerBackoff` | `"backoff"` | Em espera após falha |

---

## 9. Callbacks

### Falha em health check

```go
h := helper.New(db,
    helper.WithOnHealthCheckFailure(func(ctx context.Context, svc *domain.Service, result *domain.HealthCheckResult) error {
        // Notificar Slack, PagerDuty, etc.
        notifySlack(fmt.Sprintf("⚠️ %s está %s: %s", svc.Name, result.Status, result.Message))
        return nil
    }),
)
```

### Execução de action

```go
h := helper.New(db,
    helper.WithOnActionExecution(func(ctx context.Context, exec *domain.ActionExecution, result *domain.ActionResult) error {
        // Auditoria, logs, notificações
        audit.Log("action_executed", exec.ActionID, exec.Status)
        return nil
    }),
)
```

---

## 10. Referência da API REST

Todas as rotas são prefixadas com `/api/helper/v1`.

### Health Checks

#### `GET /api/helper/v1/healthcare/`
Retorna o status geral de todos os serviços monitorados.

**Response 200:**
```json
{
  "status": "healthy",
  "services": [
    {
      "id": "uuid",
      "service_id": "uuid",
      "status": "healthy",
      "message": "PostgreSQL connection is healthy",
      "checked_at": "2026-07-21T10:00:00Z"
    }
  ]
}
```

#### `GET /api/helper/v1/healthcare/:name`
Retorna o último status registrado de um serviço específico.

**Response 200:**
```json
{
  "id": "uuid",
  "service_id": "uuid",
  "status": "healthy",
  "message": "OK",
  "checked_at": "2026-07-21T10:00:00Z"
}
```

**Response 404:** Serviço não encontrado.

#### `POST /api/helper/v1/healthcare/:name/refresh`
Força a execução imediata do health check do serviço.

**Response 200:**
```json
{
  "status": "healthy",
  "message": "PostgreSQL connection is healthy",
  "timestamp": "2026-07-21T10:05:00Z",
  "data": {
    "open_connections": 5,
    "in_use": 2,
    "idle": 3
  }
}
```

---

### Actions

#### `GET /api/helper/v1/actions/?service_id=<uuid>`
Lista todas as actions de um serviço.

**Query param obrigatório:** `service_id`

**Response 200:**
```json
[
  {
    "id": "uuid",
    "service_id": "uuid",
    "slug": "reiniciar-servico",
    "title": "Reiniciar Serviço",
    "description": "...",
    "active": true,
    "questions": [...]
  }
]
```

#### `GET /api/helper/v1/actions/:slug`
Retorna o detalhe de uma action, incluindo suas questões.

**Response 200:**
```json
{
  "id": "uuid",
  "slug": "reiniciar-servico",
  "title": "Reiniciar Serviço",
  "questions": [
    {
      "slug": "servico",
      "input_type": "select",
      "label": "Qual serviço?",
      "required": true,
      "options": "[\"api\",\"worker\"]",
      "order": 1,
      "children": []
    }
  ]
}
```

#### `POST /api/helper/v1/actions/:slug/execute`
Executa uma action com as respostas do formulário.

**Request body:**
```json
{
  "servico": "api",
  "motivo": "Alta de memória detectada"
}
```

**Response 200:**
```json
{
  "success": true,
  "message": "Serviço reiniciado com sucesso",
  "data": {
    "servico": "api"
  }
}
```

**Response 422:** Campos obrigatórios faltando.  
**Response 404:** Action não encontrada.

---

### Workers

#### `GET /api/helper/v1/workers/`
Lista todos os workers registrados no WorkerPool.

**Response 200:**
```json
[
  {
    "id": "kafka-consumer-1",
    "name": "Kafka Consumer Principal",
    "status": "idle",
    "last_check": "2026-07-21T10:04:55Z",
    "last_error": "",
    "processed_count": 1543,
    "failed_count": 2
  }
]
```

#### `GET /api/helper/v1/workers/:id`
Retorna o detalhe de um worker específico.

**Response 200:** Mesmo formato de um item da lista acima.  
**Response 404:** Worker não encontrado.

---

## 11. Referência da API Go

### `helper.New(db *gorm.DB, opts ...Option) *Helper`
Cria uma instância do Helper.

### `h.RegisterService(name string, checker domain.HealthChecker)`
Registra um health checker para um serviço.

### `h.RegisterAction(slug string, handler domain.ActionHandler)`
Registra um handler para uma action.

### `h.RegisterRoutes(router *gin.Engine, db *gorm.DB, middleware *gin.HandlerFunc)`
Registra todas as rotas no router Gin do cliente. Passa `nil` como middleware para sem autenticação.

### `h.Sync(ctx context.Context, defs []services.ServiceDefinition) error`
Sincroniza definições de serviços e actions com o banco.

### `h.CheckService(ctx, name) (*domain.HealthCheckResult, error)`
Executa o health check de um serviço manualmente.

### `h.CheckAll(ctx) map[string]*domain.HealthCheckResult`
Executa health checks de todos os serviços registrados.

### `h.ExecuteAction(ctx, slug, input) (*domain.ActionResult, error)`
Executa uma action programaticamente (sem HTTP).

### `helper.NewScheduler(h *Helper, cfg SchedulerConfig) *Scheduler`
Cria e inicia o scheduler de health checks periódicos.

### `helper.DefaultSchedulerConfig() SchedulerConfig`
Retorna a configuração padrão (health check a cada 10 minutos).

---

## Exemplo completo de integração

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/gin-gonic/gin"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "github.com/caiofariavert/golang_vert_helper/internal/domain"
    "github.com/caiofariavert/golang_vert_helper/internal/services"
    healthchecks "github.com/caiofariavert/golang_vert_helper/pkg/health_checks"
    "github.com/caiofariavert/golang_vert_helper/pkg/helper"
)

func main() {
    // Sua conexão GORM existente
    db, _ := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})

    // WorkerPool para monitorar goroutines
    pool := healthchecks.NewWorkerPool()

    // Inicializar helper
    h := helper.New(db,
        helper.WithLogger(slog.Default()),
        helper.WithWorkerPool(pool),
        helper.WithOnHealthCheckFailure(func(ctx context.Context, svc *domain.Service, result *domain.HealthCheckResult) error {
            slog.Error("serviço com falha", "service", svc.Name, "status", result.Status)
            return nil
        }),
    )

    // Registrar health checkers
    h.RegisterService("postgres", healthchecks.NewGormPostgresChecker(db))

    // Registrar actions
    h.RegisterAction("limpar-cache", func(ctx context.Context, action *domain.Action, input map[string]interface{}) (*domain.ActionResult, error) {
        tipo := input["tipo"].(string)
        // ... lógica de limpeza
        return &domain.ActionResult{Success: true, Message: "Cache " + tipo + " limpo"}, nil
    })

    // Sincronizar definições com o banco
    h.Sync(context.Background(), []services.ServiceDefinition{
        {
            Name:        "meu-sistema",
            Description: "Backend principal",
            Actions: []services.ActionDefinition{
                {
                    Slug:  "limpar-cache",
                    Title: "Limpar Cache",
                    Questions: []services.QuestionDefinition{
                        {
                            Slug:      "tipo",
                            InputType: domain.QuestionInputTypeSelect,
                            Label:     "Tipo de cache",
                            Required:  true,
                            Options:   []string{"redis", "memcached", "local"},
                            Order:     1,
                        },
                    },
                },
            },
        },
    })

    // Iniciar scheduler (health check a cada 10 minutos)
    scheduler := helper.NewScheduler(h, helper.DefaultSchedulerConfig())
    defer scheduler.Stop()

    // Seu router Gin existente
    router := gin.Default()

    // Suas próprias rotas
    router.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"pong": true}) })

    // Middleware de autenticação (opcional)
    auth := func(c *gin.Context) {
        if c.GetHeader("X-API-Key") != os.Getenv("API_KEY") {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return
        }
        c.Next()
    }

    // Registrar rotas do helper com autenticação
    h.RegisterRoutes(router, db, &auth)

    // Registrar worker
    pool.Register("worker-principal", "Worker Principal")
    go runWorker(pool)

    router.Run(":8080")
}

func runWorker(pool *healthchecks.WorkerPool) {
    pool.UpdateStatus("worker-principal", healthchecks.WorkerRunning, "")
    for {
        pool.UpdateStatus("worker-principal", healthchecks.WorkerProcessing, "")
        // ... processar tarefa
        pool.UpdateStatus("worker-principal", healthchecks.WorkerIdle, "")
    }
}
```
