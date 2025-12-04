package handler

import "net/http"

func writeHTML(w http.ResponseWriter, status int, html string) (int, error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	status, err := w.Write([]byte(html))
	return status, err
}
