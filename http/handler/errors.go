package handler

import (
	"errors"
	"fmt"
	"net/http"
)

func (h *Handler) ErrorResponse(w http.ResponseWriter, r *http.Request, err any, status int) {
	//h.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	h.Core.Log.ErrorLog.Println(err)
	writeError(h.Core.App.Type, w, status, err)
}

func (h *Handler) InternalServerErrorResponse(w http.ResponseWriter, r *http.Request, err any) {
	//h.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	h.Core.Log.ErrorLog.Println(err)
	writeError(h.Core.App.Type, w, http.StatusInternalServerError, errors.New("the server encountered a problem"))
}

func (h *Handler) UnprocessableEntityResponse(w http.ResponseWriter, r *http.Request, err any) {
	//h.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	h.Core.Log.ErrorLog.Println(err)
	writeError(h.Core.App.Type, w, http.StatusUnprocessableEntity, err)
}

func (h *Handler) MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	//h.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeError(h.Core.App.Type, w, http.StatusMethodNotAllowed, errors.New("Method not allowed"))
}

func (h *Handler) ForbiddenResponse(w http.ResponseWriter, r *http.Request) {
	//h.logger.Warnw("forbidden", "method", r.Method, "path", r.URL.Path, "error")
	writeError(h.Core.App.Type, w, http.StatusForbidden, errors.New("forbidden"))
}

func (h *Handler) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	//h.logger.Warnf("bad request", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	h.Core.Log.ErrorLog.Println(err)
	writeError(h.Core.App.Type, w, http.StatusBadRequest, err)
}

func (h *Handler) ConflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	//h.logger.Errorf("conflict response", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	h.Core.Log.ErrorLog.Println(err)
	writeError(h.Core.App.Type, w, http.StatusConflict, err)
}

func (h *Handler) NotFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	//h.logger.Warnf("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	h.Core.Log.ErrorLog.Println(err)
	writeError(h.Core.App.Type, w, http.StatusNotFound, errors.New("not found"))
}

func (h *Handler) UnauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	//h.logger.Warnf("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	h.Core.Log.ErrorLog.Println(err)
	writeError(h.Core.App.Type, w, http.StatusUnauthorized, errors.New("unauthorized"))
}

func (h *Handler) UnauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	//h.logger.Warnf("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	h.Core.Log.ErrorLog.Println(err)
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	writeError(h.Core.App.Type, w, http.StatusUnauthorized, errors.New("unauthorized"))
}

func (h *Handler) TooManyRequestsResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	//h.logger.Warnw("rate limit exceeded", "method", r.Method, "path", r.URL.Path)
	h.Core.Log.ErrorLog.Println(retryAfter)
	w.Header().Set("Retry-After", retryAfter)

	writeError(h.Core.App.Type, w, http.StatusTooManyRequests, errors.New("rate limit exceeded, retry after: "+retryAfter))
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
