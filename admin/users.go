package admin

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/juanfont/juango/auth"
	"github.com/juanfont/juango/types"
	"github.com/rs/zerolog/log"
)

// AdminAPIHandlers provides HTTP handlers for admin user and audit log APIs.
type AdminAPIHandlers struct {
	userStore   auth.AdminUserStore
	auditStore  auth.AdminAuditStore
}

// NewAdminAPIHandlers creates new admin API handlers.
func NewAdminAPIHandlers(userStore auth.AdminUserStore, auditStore auth.AdminAuditStore) *AdminAPIHandlers {
	return &AdminAPIHandlers{
		userStore:  userStore,
		auditStore: auditStore,
	}
}

// ListUsersHandler handles GET /api/admin/users.
func (h *AdminAPIHandlers) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	search := r.URL.Query().Get("search")

	users, err := h.userStore.GetAllUsers(ctx, search)
	if err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusInternalServerError, "Failed to list users", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetUserStatsHandler handles GET /api/admin/users/stats.
func (h *AdminAPIHandlers) GetUserStatsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	total, err := h.userStore.CountUsers(ctx)
	if err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusInternalServerError, "Failed to count users", err))
		return
	}

	admins, err := h.userStore.CountAdminUsers(ctx)
	if err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusInternalServerError, "Failed to count admin users", err))
		return
	}

	active30d, err := h.userStore.CountActiveUsers(ctx, "-30 days")
	if err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusInternalServerError, "Failed to count active users", err))
		return
	}

	active7d, err := h.userStore.CountActiveUsers(ctx, "-7 days")
	if err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusInternalServerError, "Failed to count active users", err))
		return
	}

	active24h, err := h.userStore.CountActiveUsers(ctx, "-1 day")
	if err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusInternalServerError, "Failed to count active users", err))
		return
	}

	new30d, err := h.userStore.CountNewUsers(ctx, "-30 days")
	if err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusInternalServerError, "Failed to count new users", err))
		return
	}

	new7d, err := h.userStore.CountNewUsers(ctx, "-7 days")
	if err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusInternalServerError, "Failed to count new users", err))
		return
	}

	stats := types.UserStats{
		TotalUsers:     total,
		AdminUsers:     admins,
		ActiveUsers30d: active30d,
		ActiveUsers7d:  active7d,
		ActiveUsers24h: active24h,
		NewUsers30d:    new30d,
		NewUsers7d:     new7d,
		InactiveUsers:  total - active30d,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetUserHandler handles GET /api/admin/users/{id}.
func (h *AdminAPIHandlers) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusBadRequest, "Invalid user ID", err))
		return
	}

	user, err := h.userStore.GetUserByID(ctx, userID)
	if err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusNotFound, "User not found", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateUserHandler handles PATCH /api/admin/users/{id}.
func (h *AdminAPIHandlers) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	adminUser := auth.GetUserFromContext(ctx)
	vars := mux.Vars(r)

	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusBadRequest, "Invalid user ID", err))
		return
	}

	var req types.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusBadRequest, "Invalid request body", err))
		return
	}

	user, err := h.userStore.GetUserByID(ctx, userID)
	if err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusNotFound, "User not found", err))
		return
	}

	// Track changes for audit log
	changes := make(map[string]interface{})

	if req.IsAdmin != nil && *req.IsAdmin != user.IsAdmin {
		// Prevent removing admin from the last admin
		if !*req.IsAdmin {
			adminCount, err := h.userStore.CountAdminUsers(ctx)
			if err != nil {
				types.WriteHTTPError(w, types.NewHTTPError(http.StatusInternalServerError, "Failed to count admins", err))
				return
			}
			if adminCount <= 1 {
				types.WriteHTTPError(w, types.NewHTTPError(http.StatusBadRequest, "Cannot remove the last admin", nil))
				return
			}
		}
		changes["is_admin"] = map[string]interface{}{"from": user.IsAdmin, "to": *req.IsAdmin}
		user.IsAdmin = *req.IsAdmin
	}

	if req.Name != nil && *req.Name != user.Name {
		changes["name"] = map[string]interface{}{"from": user.Name, "to": *req.Name}
		user.Name = *req.Name
	}

	if req.Active != nil {
		isActive := user.IsActive()
		if *req.Active && !isActive {
			// Reactivate
			user.DeletedAt.Valid = false
			changes["active"] = map[string]interface{}{"from": false, "to": true}
		} else if !*req.Active && isActive {
			// Deactivate
			user.DeletedAt.Valid = true
			changes["active"] = map[string]interface{}{"from": true, "to": false}
		}
	}

	if err := h.userStore.UpdateUser(ctx, user); err != nil {
		types.WriteHTTPError(w, types.NewHTTPError(http.StatusInternalServerError, "Failed to update user", err))
		return
	}

	// Audit log
	if h.auditStore != nil && len(changes) > 0 {
		action := types.ActionUserUpdated
		if _, ok := changes["is_admin"]; ok {
			if user.IsAdmin {
				action = types.ActionUserAdminGranted
			} else {
				action = types.ActionUserAdminRevoked
			}
		}
		if activeChange, ok := changes["active"]; ok {
			ac := activeChange.(map[string]interface{})
			if ac["to"] == false {
				action = types.ActionUserDeactivated
			} else {
				action = types.ActionUserReactivated
			}
		}

		auditLog := auth.NewAuditLogWithContext(
			ctx,
			action,
			types.ResourceTypeUser,
			user.ID.String(),
		).WithChanges(changes).
			WithIPAddress(auth.GetClientIP(r)).
			WithUserAgent(r.UserAgent())
		auditLog.AddDetail("admin_email", adminUser.Email)

		if err := h.auditStore.CreateAuditLog(ctx, auditLog); err != nil {
			log.Error().Err(err).Msg("Failed to create audit log for user update")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
