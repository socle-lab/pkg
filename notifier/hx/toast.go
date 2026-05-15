package hx

import (
	"encoding/json"

	"github.com/socle-lab/pkg/notifier"
)

type Toast struct {
	Message  string         `json:"message"`
	Level    notifier.Level `json:"level"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

func (t Toast) MarshalJSON() ([]byte, error) {
	type out struct {
		Message  jsonStringASCII `json:"message"`
		Level    notifier.Level  `json:"level"`
		Metadata map[string]any  `json:"metadata,omitempty"`
	}
	return json.Marshal(out{
		Message:  jsonStringASCII(t.Message),
		Level:    t.Level,
		Metadata: t.Metadata,
	})
}

func (n *HTMXNotifier) Toast(msg string, level notifier.Level) {
	n.events = append(n.events, notifier.Event{
		Type: "socle:toast",
		Data: Toast{
			Message: msg,
			Level:   level,
		},
	})
}

func (n *HTMXNotifier) ToastWithMetadata(msg string, level notifier.Level, metadata map[string]any) {
	n.events = append(n.events, notifier.Event{
		Type: "socle:toast",
		Data: Toast{
			Message:  msg,
			Level:    level,
			Metadata: metadata,
		},
	})
}
