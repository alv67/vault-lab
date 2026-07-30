package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/amelamela/vault-lab/internal/auth"
	"github.com/amelamela/vault-lab/internal/model"
	"github.com/amelamela/vault-lab/internal/service"
)

type Handler struct {
	svc     *service.Service
	jwtAuth *auth.JWTAuth
}

func New(svc *service.Service, jwtAuth *auth.JWTAuth) *Handler {
	return &Handler{svc: svc, jwtAuth: jwtAuth}
}

func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respond(w, status, map[string]string{"error": msg})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.svc.Register(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		if err == service.ErrEmailExists {
			respondError(w, http.StatusConflict, "email already registered")
			return
		}
		log.Error().Err(err).Msg("registration failed")
		respondError(w, http.StatusInternalServerError, "registration failed")
		return
	}

	respond(w, http.StatusCreated, user)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, accessToken, refreshToken, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			respondError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		log.Error().Err(err).Msg("login failed")
		respondError(w, http.StatusInternalServerError, "login failed")
		return
	}

	respond(w, http.StatusOK, map[string]interface{}{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	accessToken, refreshToken, err := h.svc.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	respond(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.svc.GetCurrentUser(r.Context(), claims)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respond(w, http.StatusOK, user)
}

func (h *Handler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	respondError(w, http.StatusNotImplemented, "not implemented yet")
}

func (h *Handler) SearchAssets(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "query parameter required")
		return
	}

	assets, err := h.svc.SearchAssets(r.Context(), query)
	if err != nil {
		log.Error().Err(err).Msg("search assets failed")
		respondError(w, http.StatusInternalServerError, "search failed")
		return
	}

	respond(w, http.StatusOK, assets)
}

func (h *Handler) LookupAsset(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "query parameter required")
		return
	}

	results, err := h.svc.LookupAsset(r.Context(), query)
	if err != nil {
		log.Error().Err(err).Msg("lookup asset failed")
		respondError(w, http.StatusInternalServerError, "lookup failed")
		return
	}

	respond(w, http.StatusOK, results)
}

func (h *Handler) ListAssets(w http.ResponseWriter, r *http.Request) {
	assets, err := h.svc.ListAssets(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("list assets failed")
		respondError(w, http.StatusInternalServerError, "list failed")
		return
	}
	respond(w, http.StatusOK, assets)
}

func (h *Handler) GetAsset(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	uid, err := parseUUID(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid asset id")
		return
	}

	asset, err := h.svc.GetAsset(r.Context(), uid)
	if err != nil {
		respondError(w, http.StatusNotFound, "asset not found")
		return
	}

	respond(w, http.StatusOK, asset)
}

func (h *Handler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	var asset model.Asset
	if err := json.NewDecoder(r.Body).Decode(&asset); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.svc.CreateAsset(r.Context(), &asset)
	if err != nil {
		log.Error().Err(err).Msg("create asset failed")
		respondError(w, http.StatusInternalServerError, "create failed")
		return
	}

	respond(w, http.StatusCreated, created)
}
