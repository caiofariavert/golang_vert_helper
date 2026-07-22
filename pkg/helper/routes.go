package helper

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/caiofariavert/golang_vert_helper/internal/adapters"
)

// RegisterRoutes registra todas as rotas do Vert Helper no router Gin do cliente.
//
// Parâmetros:
//   - router: o *gin.Engine já configurado pelo cliente
//   - db: conexão GORM para uso dos repositories
//   - middleware: middleware opcional (ex: autenticação). Passa nil para sem middleware.
//
// Rotas registradas em /api/helper/v1/:
//
//	GET  /healthcare/                     → status geral de todos os serviços
//	GET  /healthcare/:name                → status de um serviço específico (query opcional: force_refresh=true)
//	GET  /actions/                        → lista actions (query: service_id)
//	GET  /actions/:slug                   → detalhe de uma action
//	POST /actions/:slug/execute           → executa uma action
//	GET  /workers/                        → lista workers do WorkerPool
//	GET  /workers/:id                     → detalhe de um worker
func (h *Helper) RegisterRoutes(router *gin.Engine, db *gorm.DB, middleware *gin.HandlerFunc) {
	handlers := adapters.NewHandlers(h.healthService, h.actionService, h.workerPool)

	group := router.Group("/api/helper/v1")

	if middleware != nil {
		group.Use(*middleware)
	}

	// Health check
	group.GET("/healthcare/", handlers.GetHealthcare)
	group.GET("/healthcare/:name", handlers.GetServiceHealth)

	// Actions
	group.GET("/actions/", handlers.ListActions)
	group.GET("/actions/:slug", handlers.GetAction)
	group.POST("/actions/:slug/execute", handlers.ExecuteAction)

	// Workers
	group.GET("/workers/", handlers.ListWorkers)
	group.GET("/workers/:id", handlers.GetWorker)
}
