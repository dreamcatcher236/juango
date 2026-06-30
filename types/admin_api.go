package types

import (
	"time"
)

// AuditLogFilter provides filtering parameters for audit log queries.
type AuditLogFilter struct {
	ActorUserID  string     `json:"actor_user_id,omitempty"`
	Action       string     `json:"action,omitempty"`
	ResourceType string     `json:"resource_type,omitempty"`
	ResourceID   string     `json:"resource_id,omitempty"`
	Search       string     `json:"search,omitempty"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}

// UserStats holds statistics about users in the system.
type UserStats struct {
	TotalUsers      int `json:"total_users"`
	AdminUsers      int `json:"admin_users"`
	ActiveUsers30d  int `json:"active_users_30d"`
	ActiveUsers7d   int `json:"active_users_7d"`
	ActiveUsers24h  int `json:"active_users_24h"`
	NewUsers30d     int `json:"new_users_30d"`
	NewUsers7d      int `json:"new_users_7d"`
	InactiveUsers   int `json:"inactive_users"`
}

// AuditLogStats holds statistics about audit log entries.
type AuditLogStats struct {
	TotalEntries     int      `json:"total_entries"`
	EntriesToday     int      `json:"entries_today"`
	UniqueActions    int      `json:"unique_actions"`
	TopActions       []string `json:"top_actions,omitempty"`
	TopResourceTypes []string `json:"top_resource_types,omitempty"`
}

// UpdateUserRequest is the request body for updating a user via the admin API.
type UpdateUserRequest struct {
	IsAdmin *bool   `json:"is_admin,omitempty"`
	Name    *string `json:"name,omitempty"`
	Active  *bool   `json:"active,omitempty"`
}

// HealthResponse is the response from the health check endpoint.
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}
