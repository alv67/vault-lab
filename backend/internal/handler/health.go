package handler

import (
	"encoding/json"
	"net/http"

	"github.com/amelamela/vault-lab/internal/service"
)

type HealthHandler struct {
	svc *service.HealthService
}

func NewHealthHandler(svc *service.HealthService) *HealthHandler {
	return &HealthHandler{svc: svc}
}

func (h *HealthHandler) GetPriceHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	summary, events, err := h.svc.GetPriceHealth(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"summary": summary,
		"events":  events,
	})
}
