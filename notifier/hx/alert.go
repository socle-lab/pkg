package hx

import "github.com/socle-lab/pkg/notifier"

type Alert struct {
	Title    string         `json:"title"`
	Message  string         `json:"message"`
	Level    notifier.Level `json:"level"`
	Metadata map[string]any `json:"metadata"`
}

func (n *HTMXNotifier) Alert(title, message string, level notifier.Level) {
	n.events = append(n.events, notifier.Event{
		Type: "socle:alert",
		Data: Alert{
			Title:   title,
			Message: message,
			Level:   level,
		},
	})
}

func (n *HTMXNotifier) AlertWithMetadata(title, message string, level notifier.Level, metadata map[string]any) {
	n.events = append(n.events, notifier.Event{
		Type: "socle:alert",
		Data: Alert{
			Title:    title,
			Message:  message,
			Level:    level,
			Metadata: metadata,
		},
	})
}
