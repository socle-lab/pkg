package hx

import (
	"encoding/json"
	"net/http"

	"github.com/socle-lab/pkg/notifier"
)

// hxTriggerPayload groups events by Type for HTMX HX-Trigger: each key is the custom
// event name; the value is a single payload object, or a slice when several events
// share the same Type.
func hxTriggerPayload(events []notifier.Event) map[string]any {
	grouped := make(map[string][]any)
	for _, e := range events {
		grouped[e.Type] = append(grouped[e.Type], e.Data)
	}
	out := make(map[string]any, len(grouped))
	for k, list := range grouped {
		if len(list) == 1 {
			out[k] = list[0]
		} else {
			out[k] = list
		}
	}
	return out
}

type HTMXNotifier struct {
	events []notifier.Event
}

func NewNotifier() notifier.Notifier {
	return &HTMXNotifier{
		events: []notifier.Event{},
	}
}

func (n *HTMXNotifier) Add(event notifier.Event) {
	n.events = append(n.events, event)
}

func (n *HTMXNotifier) Notify(w http.ResponseWriter) {
	payload := hxTriggerPayload(n.events)
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	// payload := map[string]interface{}{
	// 	"socle:event": n.events,
	// }

	//data, _ := json.Marshal(payload)

	w.Header().Set("HX-Trigger", string(data))
}
