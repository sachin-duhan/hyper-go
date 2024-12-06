package models

import (
	"encoding/json"
	"time"
)

type AnalyticsEvent struct {
	ID         string            `json:"id,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
	UserID     uint              `json:"user_id"`
	Event      string            `json:"event"`
	Metadata   string            `json:"metadata,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
}

// Common analytics events
const (
	EventPageView     = "page_view"
	EventButtonClick  = "button_click"
	EventFormSubmit   = "form_submit"
	EventAPIRequest   = "api_request"
	EventUserSignup   = "user_signup"
	EventUserLogin    = "user_login"
	EventUserLogout   = "user_logout"
	EventErrorOccured = "error_occurred"
)

// Helper method to set metadata as JSON
func (e *AnalyticsEvent) SetMetadata(data interface{}) error {
	if data == nil {
		e.Metadata = ""
		return nil
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	e.Metadata = string(bytes)
	return nil
}

// Helper method to get metadata as JSON
func (e *AnalyticsEvent) GetMetadata(out interface{}) error {
	if e.Metadata == "" {
		return nil
	}

	return json.Unmarshal([]byte(e.Metadata), out)
}
