package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/vert/golang_vert_helper/internal/domain"
)

// ActionService gerencia o ciclo de vida das actions
type ActionService struct {
	actionRepo    domain.ActionRepository
	executionRepo domain.ActionExecutionRepository
	handlers      map[string]domain.ActionHandler
	onExecution   domain.OnActionExecution
	logger        *slog.Logger
}

// NewActionService cria um novo ActionService
func NewActionService(
	actionRepo domain.ActionRepository,
	executionRepo domain.ActionExecutionRepository,
	logger *slog.Logger,
) *ActionService {
	if logger == nil {
		logger = slog.Default()
	}
	return &ActionService{
		actionRepo:    actionRepo,
		executionRepo: executionRepo,
		handlers:      make(map[string]domain.ActionHandler),
		logger:        logger,
	}
}

// RegisterHandler registra um handler para um slug de ação
func (s *ActionService) RegisterHandler(slug string, handler domain.ActionHandler) {
	s.handlers[slug] = handler
}

// SetOnExecution define o callback chamado após cada execução
func (s *ActionService) SetOnExecution(fn domain.OnActionExecution) {
	s.onExecution = fn
}

// GetAction retorna uma ação pelo slug
func (s *ActionService) GetAction(ctx context.Context, slug string) (*domain.Action, error) {
	return s.actionRepo.GetBySlug(ctx, slug)
}

// ListActions retorna todas as ações de um serviço
func (s *ActionService) ListActions(ctx context.Context, serviceID string) ([]*domain.Action, error) {
	return s.actionRepo.ListByServiceID(ctx, serviceID)
}

// Execute executa uma ação com as respostas fornecidas pelo usuário
func (s *ActionService) Execute(ctx context.Context, slug string, input map[string]interface{}) (*domain.ActionResult, error) {
	action, err := s.actionRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if !action.Active {
		return nil, fmt.Errorf("%w: %s is inactive", domain.ErrActionNotFound, slug)
	}

	handler, ok := s.handlers[slug]
	if !ok {
		return nil, fmt.Errorf("%w: no handler registered for %s", domain.ErrActionNotFound, slug)
	}

	// Valida as respostas contra as questões da action
	if err := s.validateInput(action, input); err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrInvalidResponses, err.Error())
	}

	// Persiste execução como "running"
	execution := &domain.ActionExecution{
		ID:       uuid.New().String(),
		ActionID: action.ID,
		Status:   domain.ActionExecutionStatusRunning,
		Input:    marshalJSON(input),
	}
	if err := s.executionRepo.Create(ctx, execution); err != nil {
		s.logger.Error("failed to create execution record", "action", slug, "error", err)
	}

	// Executa o handler
	result, handlerErr := handler(ctx, action, input)

	// Atualiza execução com o resultado
	now := time.Now()
	if handlerErr != nil {
		execution.Status = domain.ActionExecutionStatusFailed
		execution.Error = handlerErr.Error()
		result = &domain.ActionResult{
			Success: false,
			Error:   handlerErr.Error(),
		}
	} else {
		execution.Status = domain.ActionExecutionStatusSuccess
		execution.Output = marshalJSON(result.Data)
	}

	execution.ExecutedAt.Time = now
	execution.ExecutedAt.Valid = true

	if err := s.executionRepo.Update(ctx, execution); err != nil {
		s.logger.Error("failed to update execution record", "action", slug, "error", err)
	}

	// Dispara callback
	if s.onExecution != nil {
		if cbErr := s.onExecution(ctx, execution, result); cbErr != nil {
			s.logger.Error("execution callback returned error", "action", slug, "error", cbErr)
		}
	}

	s.logger.Info("action executed",
		"action", slug,
		"success", result.Success,
	)

	return result, handlerErr
}

// validateInput verifica se todas as questões obrigatórias foram respondidas
func (s *ActionService) validateInput(action *domain.Action, input map[string]interface{}) error {
	for _, q := range action.Questions {
		if !q.Required {
			continue
		}

		// Questões com parent só são obrigatórias se o parent tiver o valor esperado
		if q.ParentID != nil && q.ParentValue != "" {
			parentVal, ok := input[*q.ParentID]
			if !ok || fmt.Sprintf("%v", parentVal) != q.ParentValue {
				continue
			}
		}

		val, ok := input[q.Slug]
		if !ok || val == nil || val == "" {
			return fmt.Errorf("required field missing: %s", q.Slug)
		}
	}
	return nil
}

// marshalJSON converte um valor para JSON string, retornando vazio em caso de erro
func marshalJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
