# Vert Helper Go

Biblioteca Go para monitoramento de saúde de serviços e execução de ações interativas.

📖 **[Ver Manual de Uso completo](docs/manual_de_uso.md)**

---

## Instalação rápida

```bash
go get github.com/caiofariavert/golang_vert_helper
```

## Exemplo mínimo

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/caiofariavert/golang_vert_helper/pkg/helper"
    healthchecks "github.com/caiofariavert/golang_vert_helper/pkg/health_checks"
)

h := helper.New(db)
h.RegisterService("postgres", healthchecks.NewGormPostgresChecker(db))

router := gin.Default()
h.RegisterRoutes(router, db, nil)
router.Run(":8080")
```

Para documentação completa, exemplos avançados e referência da API, consulte o **[Manual de Uso](docs/manual_de_uso.md)**.

## Importante para integração externa

Ao usar esta biblioteca em outro projeto, **nao importe pacotes `internal/*`**.
Use os tipos publicos em `pkg/contracts` quando precisar de structs, enums e definicoes de sync.

Exemplo:

```go
import "github.com/caiofariavert/golang_vert_helper/pkg/contracts"
```
