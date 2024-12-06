package models

import "time"

type AuditLog struct {
	ID         string    `json:"id,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	UserID     uint64    `json:"user_id"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id"`
	Details    string    `json:"details,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
}

// Common audit actions
const (
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionLogin  = "login"
	ActionLogout = "logout"
)

// Common resources
const (
	ResourceUser    = "user"
	ResourceProfile = "profile"
	ResourcePost    = "post"
	ResourceComment = "comment"
)
