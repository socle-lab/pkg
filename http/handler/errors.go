package handler

import (
	"errors"
	"fmt"
	"net/http"
)

func (h *Handler) Error(w http.ResponseWriter, r *http.Request, err any, status int) {
	//h.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	h.Core.Log.ErrorLog.Println("internal error", "method", r.Method, "path", r.URL.Path, "error", err)
	writeError(h.Core.AppModule.Type, w, status, err)
}

func (h *Handler) InternalServerError(w http.ResponseWriter, r *http.Request, err any) {
	//h.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	h.Core.Log.ErrorLog.Println(err)
	writeError(h.Core.AppModule.Type, w, http.StatusInternalServerError, errors.New("the server encountered a problem"))
}

func (h *Handler) UnprocessableEntity(w http.ResponseWriter, r *http.Request, err any) {
	//h.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	h.Core.Log.ErrorLog.Println(err)
	writeError(h.Core.AppModule.Type, w, http.StatusUnprocessableEntity, err)
}

func (h *Handler) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	//h.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeError(h.Core.AppModule.Type, w, http.StatusMethodNotAllowed, errors.New("Method not allowed"))
}

func (h *Handler) Forbidden(w http.ResponseWriter, r *http.Request) {
	//h.logger.Warnw("forbidden", "method", r.Method, "path", r.URL.Path, "error")
	writeError(h.Core.AppModule.Type, w, http.StatusForbidden, errors.New("forbidden"))
}

func (h *Handler) BadRequest(w http.ResponseWriter, r *http.Request, err error) {
	//h.logger.Warnf("bad request", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	h.Core.Log.ErrorLog.Println(err)
	writeError(h.Core.AppModule.Type, w, http.StatusBadRequest, err)
}

func (h *Handler) Conflict(w http.ResponseWriter, r *http.Request, err error) {
	//h.logger.Errorf("conflict response", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	h.Core.Log.ErrorLog.Println(err)
	writeError(h.Core.AppModule.Type, w, http.StatusConflict, err)
}

func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request, err error) {
	//h.logger.Warnf("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	h.Core.Log.ErrorLog.Println(err)
	writeError(h.Core.AppModule.Type, w, http.StatusNotFound, errors.New("not found"))
}

func (h *Handler) Unauthorized(w http.ResponseWriter, r *http.Request, err error) {
	//h.logger.Warnf("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	h.Core.Log.ErrorLog.Println(err)
	writeError(h.Core.AppModule.Type, w, http.StatusUnauthorized, errors.New("unauthorized"))
}

func (h *Handler) UnauthorizedBasic(w http.ResponseWriter, r *http.Request, err error) {
	//h.logger.Warnf("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	h.Core.Log.ErrorLog.Println(err)
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	writeError(h.Core.AppModule.Type, w, http.StatusUnauthorized, errors.New("unauthorized"))
}

func (h *Handler) TooManyRequests(w http.ResponseWriter, r *http.Request, retryAfter string) {
	//h.logger.Warnw("rate limit exceeded", "method", r.Method, "path", r.URL.Path)
	h.Core.Log.ErrorLog.Println(retryAfter)
	w.Header().Set("Retry-After", retryAfter)

	writeError(h.Core.AppModule.Type, w, http.StatusTooManyRequests, errors.New("rate limit exceeded, retry after: "+retryAfter))
}

func writeError(moduleType string, w http.ResponseWriter, statusCode int, err any) {

	if moduleType == "web" {
		if err, ok := err.(error); ok {
			http.Error(w, err.Error(), statusCode)
		} else {
			writeHTML(w, statusCode, fmt.Sprintf("%v", err))
		}
	} else {
		if err, ok := err.(error); ok {
			writeJSONError(w, statusCode, err.Error())
		} else {
			writeJSON(w, statusCode, err)
		}

	}
}
