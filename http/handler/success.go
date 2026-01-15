package handler

import "net/http"

func (h *Handler) OK(w http.ResponseWriter, data any) error {
	type envelope struct {
		Data any `json:"data"`
	}

	return writeJSON(w, http.StatusOK, &envelope{Data: data})
}
