package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/amelamela/vault-lab/internal/auth"
	"github.com/amelamela/vault-lab/internal/model"
	"github.com/amelamela/vault-lab/internal/service"
)

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func (h *Handler) CreatePortfolio(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Currency    string `json:"currency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}

	portfolio, err := h.svc.CreatePortfolio(r.Context(), claims.UserID, req.Name, req.Description, req.Currency)
	if err != nil {
		log.Error().Err(err).Msg("create portfolio failed")
		respondError(w, http.StatusInternalServerError, "create failed")
		return
	}

	respond(w, http.StatusCreated, portfolio)
}

func (h *Handler) ListPortfolios(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	portfolios, err := h.svc.ListPortfolios(r.Context(), claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("list portfolios failed")
		respondError(w, http.StatusInternalServerError, "list failed")
		return
	}

	respond(w, http.StatusOK, portfolios)
}

func (h *Handler) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	uid, err := parseUUID(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid portfolio id")
		return
	}

	portfolio, err := h.svc.GetPortfolio(r.Context(), uid, claims.UserID)
	if err != nil {
		if err == service.ErrForbidden {
			respondError(w, http.StatusForbidden, "forbidden")
			return
		}
		respondError(w, http.StatusNotFound, "portfolio not found")
		return
	}

	respond(w, http.StatusOK, portfolio)
}

func (h *Handler) UpdatePortfolio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	uid, err := parseUUID(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid portfolio id")
		return
	}

	var p model.Portfolio
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	p.ID = uid

	if err := h.svc.UpdatePortfolio(r.Context(), &p); err != nil {
		log.Error().Err(err).Msg("update portfolio failed")
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}

	respond(w, http.StatusOK, p)
}

func (h *Handler) DeletePortfolio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	uid, err := parseUUID(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid portfolio id")
		return
	}

	if err := h.svc.DeletePortfolio(r.Context(), uid); err != nil {
		log.Error().Err(err).Msg("delete portfolio failed")
		respondError(w, http.StatusInternalServerError, "delete failed")
		return
	}

	respond(w, http.StatusNoContent, nil)
}

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	portfolioID, err := parseUUID(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid portfolio id")
		return
	}

	var tx model.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	tx.PortfolioID = portfolioID

	created, err := h.svc.AddTransaction(r.Context(), &tx)
	if err != nil {
		log.Error().Err(err).Msg("create transaction failed")
		respondError(w, http.StatusInternalServerError, "create failed")
		return
	}

	respond(w, http.StatusCreated, created)
}

func (h *Handler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	txID, err := parseUUID(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	var tx model.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	tx.ID = txID

	if err := h.svc.UpdateTransaction(r.Context(), claims.UserID, &tx); err != nil {
		if err == service.ErrForbidden {
			respondError(w, http.StatusForbidden, "forbidden")
			return
		}
		if err == service.ErrNotFound {
			respondError(w, http.StatusNotFound, "transaction not found")
			return
		}
		log.Error().Err(err).Msg("update transaction failed")
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}

	respond(w, http.StatusOK, tx)
}

func (h *Handler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	txID, err := parseUUID(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	if err := h.svc.DeleteTransaction(r.Context(), claims.UserID, txID); err != nil {
		if err == service.ErrForbidden {
			respondError(w, http.StatusForbidden, "forbidden")
			return
		}
		if err == service.ErrNotFound {
			respondError(w, http.StatusNotFound, "transaction not found")
			return
		}
		log.Error().Err(err).Msg("delete transaction failed")
		respondError(w, http.StatusInternalServerError, "delete failed")
		return
	}

	respond(w, http.StatusNoContent, nil)
}

func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	portfolioID, err := parseUUID(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid portfolio id")
		return
	}

	txs, err := h.svc.ListTransactions(r.Context(), portfolioID)
	if err != nil {
		log.Error().Err(err).Msg("list transactions failed")
		respondError(w, http.StatusInternalServerError, "list failed")
		return
	}

	respond(w, http.StatusOK, txs)
}

func (h *Handler) GetPortfolioSummary(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	portfolioID, err := parseUUID(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid portfolio id")
		return
	}

	if _, err := h.svc.GetPortfolio(r.Context(), portfolioID, claims.UserID); err != nil {
		if err == service.ErrForbidden {
			respondError(w, http.StatusForbidden, "forbidden")
			return
		}
		respondError(w, http.StatusNotFound, "portfolio not found")
		return
	}

	summary, err := h.svc.GetPortfolioSummary(r.Context(), portfolioID)
	if err != nil {
		log.Error().Err(err).Msg("get summary failed")
		respondError(w, http.StatusInternalServerError, "summary failed")
		return
	}

	respond(w, http.StatusOK, summary)
}

func (h *Handler) GetPortfolioAllocation(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	portfolioID, err := parseUUID(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid portfolio id")
		return
	}

	if _, err := h.svc.GetPortfolio(r.Context(), portfolioID, claims.UserID); err != nil {
		if err == service.ErrForbidden {
			respondError(w, http.StatusForbidden, "forbidden")
			return
		}
		respondError(w, http.StatusNotFound, "portfolio not found")
		return
	}

	allocation, err := h.svc.GetPortfolioAllocation(r.Context(), portfolioID)
	if err != nil {
		log.Error().Err(err).Msg("get allocation failed")
		respondError(w, http.StatusInternalServerError, "allocation failed")
		return
	}

	respond(w, http.StatusOK, allocation)
}

func (h *Handler) GetPortfolioPerformance(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	portfolioID, err := parseUUID(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid portfolio id")
		return
	}

	if _, err := h.svc.GetPortfolio(r.Context(), portfolioID, claims.UserID); err != nil {
		if err == service.ErrForbidden {
			respondError(w, http.StatusForbidden, "forbidden")
			return
		}
		respondError(w, http.StatusNotFound, "portfolio not found")
		return
	}

	performance, err := h.svc.GetPortfolioPerformance(r.Context(), portfolioID)
	if err != nil {
		log.Error().Err(err).Msg("get performance failed")
		respondError(w, http.StatusInternalServerError, "performance failed")
		return
	}

	respond(w, http.StatusOK, performance)
}

func (h *Handler) GetPortfolioROI(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	portfolioID, err := parseUUID(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid portfolio id")
		return
	}

	if _, err := h.svc.GetPortfolio(r.Context(), portfolioID, claims.UserID); err != nil {
		if err == service.ErrForbidden {
			respondError(w, http.StatusForbidden, "forbidden")
			return
		}
		respondError(w, http.StatusNotFound, "portfolio not found")
		return
	}

	roi, err := h.svc.GetPortfolioROI(r.Context(), portfolioID)
	if err != nil {
		log.Error().Err(err).Msg("get roi failed")
		respondError(w, http.StatusInternalServerError, "roi failed")
		return
	}

	respond(w, http.StatusOK, roi)
}

func (h *Handler) GetPortfolioHistory(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	uid, err := parseUUID(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid portfolio id")
		return
	}

	if _, err := h.svc.GetPortfolio(r.Context(), uid, claims.UserID); err != nil {
		if err == service.ErrForbidden {
			respondError(w, http.StatusForbidden, "forbidden")
			return
		}
		respondError(w, http.StatusNotFound, "portfolio not found")
		return
	}

	history, err := h.svc.GetPortfolioHistory(r.Context(), uid)
	if err != nil {
		log.Error().Err(err).Msg("get portfolio history failed")
		respondError(w, http.StatusInternalServerError, "history failed")
		return
	}

	respond(w, http.StatusOK, history)
}

func (h *Handler) GetPrices(w http.ResponseWriter, r *http.Request) {
	assetID := chi.URLParam(r, "assetID")
	uid, err := parseUUID(assetID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid asset id")
		return
	}

	prices, err := h.svc.GetPrices(r.Context(), uid)
	if err != nil {
		log.Error().Err(err).Msg("get prices failed")
		respondError(w, http.StatusInternalServerError, "get prices failed")
		return
	}

	respond(w, http.StatusOK, prices)
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	dash, err := h.svc.GetDashboard(r.Context(), claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("get dashboard failed")
		respondError(w, http.StatusInternalServerError, "dashboard failed")
		return
	}

	respond(w, http.StatusOK, dash)
}

func (h *Handler) RefreshPrices(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var portfolioID *uuid.UUID
	if pid := r.URL.Query().Get("portfolio_id"); pid != "" {
		uid, err := parseUUID(pid)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid portfolio id")
			return
		}
		if _, err := h.svc.GetPortfolio(r.Context(), uid, claims.UserID); err != nil {
			respondError(w, http.StatusForbidden, "forbidden")
			return
		}
		portfolioID = &uid
	}

	report, err := h.svc.RefreshPrices(r.Context(), portfolioID)
	if err != nil {
		log.Error().Err(err).Msg("refresh prices failed")
		respondError(w, http.StatusInternalServerError, "refresh failed")
		return
	}

	respond(w, http.StatusOK, report)
}
