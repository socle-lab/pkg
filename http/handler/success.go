package handler

import "net/http"

func (h *Handler) OK(w http.ResponseWriter, data any) error {
	return writeJSON(w, http.StatusOK, data)
}
