package admin

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/juanfont/juango/types"
)

// DBPinger is an interface for database health checking.
type DBPinger interface {
	PingContext(ctx context.Context) error
}

// HealthHandler provides the health check endpoint.
type HealthHandler struct {
	db DBPinger
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(db DBPinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// ServeHTTP handles GET /api/health.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dbStatus := "ok"
	if err := h.db.PingContext(ctx); err != nil {
		dbStatus = "error"
	}

	status := "ok"
	if dbStatus != "ok" {
		status = "degraded"
	}

	resp := types.HealthResponse{
		Status:   status,
		Database: dbStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	if status != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(resp)
}
