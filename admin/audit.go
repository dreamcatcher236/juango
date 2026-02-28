package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/juanfont/juango/auth"
	"github.com/juanfont/juango/types"
)

// ListAuditLogsHandler handles GET /api/admin/audit-logs.
func (h *AdminAPIHandlers) ListAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	filter := types.AuditLogFilter{
		ActorUserID:  q.Get("actor_user_id"),
		Action:       q.Get("action"),
		ResourceType: q.Get("resource_type"),
		ResourceID:   q.Get("resource_id"),
		Search:       q.Get("search"),
	}

	if v := q.Get("start_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.StartTime = &t
		}
	}
	if v := q.Get("end_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.EndTime = &t
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Offset = n
		}
	}

	if filter.Limit <= 0 || filter.Limit > 500 {
		filter.Limit = 50
	}

	logs, err := h.auditStore.GetAuditLogs(ctx, filter)
	if err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusInternalServerError, "Failed to list audit logs", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// GetAuditLogStatsHandler handles GET /api/admin/audit-logs/stats.
func (h *AdminAPIHandlers) GetAuditLogStatsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	totalFilter := types.AuditLogFilter{}
	total, err := h.auditStore.CountAuditLogs(ctx, totalFilter)
	if err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusInternalServerError, "Failed to count audit logs", err))
		return
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayFilter := types.AuditLogFilter{StartTime: &todayStart}
	today, err := h.auditStore.CountAuditLogs(ctx, todayFilter)
	if err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusInternalServerError, "Failed to count today's audit logs", err))
		return
	}

	actions, err := h.auditStore.GetDistinctActions(ctx)
	if err != nil {
		actions = []string{}
	}

	resourceTypes, err := h.auditStore.GetDistinctResourceTypes(ctx)
	if err != nil {
		resourceTypes = []string{}
	}

	_ = auth.GetUserFromContext(ctx) // ensure user context exists

	stats := types.AuditLogStats{
		TotalEntries:     total,
		EntriesToday:     today,
		UniqueActions:    len(actions),
		TopActions:       actions,
		TopResourceTypes: resourceTypes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
