// Package notifier defines a small abstraction for queuing typed events during a
// request and writing them to the HTTP response (e.g. HTMX HX-Trigger JSON).
package notifier

import "net/http"

// Level classifies toast/alert styling or severity for the client (theme, icon, etc.).
type Level string

const (
	Success Level = "success"
	Error   Level = "error"
	Warning Level = "warning"
	Info    Level = "info"
)

// Event is a generic named payload the client can subscribe to (custom events, etc.).
// Type is the event name; Data is optional structured detail for that handler.
type Event struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

// Notifier collects client-side events for the current request; Notify writes them to w.
// Implementations may map Toast and Alert to Event entries with conventional Type/Data.
type Notifier interface {
	Add(event Event)

	Toast(message string, level Level)
	ToastWithMetadata(message string, level Level, metadata map[string]any)

	Alert(title, message string, level Level)
	AlertWithMetadata(title, message string, level Level, metadata map[string]any)

	Notify(w http.ResponseWriter)
}
