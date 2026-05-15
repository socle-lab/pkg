package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/rpc"
	"reflect"
	"strings"

	"github.com/socle-lab/core"
)

// Handlers is the type for handlers, and gives access to Socle and models
type Handler struct {
	Core *core.Core
}

// isRPCConnectionClosed reports typical errors when the remote RPC server is down or the TCP link died.
func isRPCConnectionClosed(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, rpc.ErrShutdown) {
		return true
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "connection is shut down") ||
		strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "connection reset") ||
		strings.Contains(s, "use of closed network connection")
}

// CallRPC runs an RPC call when the client is connected; otherwise responds with 500 and returns false.
func (h *Handler) CallRPC(w http.ResponseWriter, r *http.Request, service *rpc.Client, serviceMethod string, args any, reply any) error {
	if service == nil {
		return errors.New("RPC client not connected")
	}
	if reply == nil {
		return errors.New("RPC reply must be a non-nil pointer")
	}
	// net/rpc uses gob; gob cannot encode a nil interface value (or a typed-nil held in an interface).
	// When callers have no arguments, pass an empty struct instead of nil.
	if args == nil {
		args = struct{}{}
	} else {
		v := reflect.ValueOf(args)
		switch v.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
			if v.IsNil() {
				args = struct{}{}
			}
		}
	}
	if err := service.Call(serviceMethod, args, reply); err != nil {
		if isRPCConnectionClosed(err) {
			return fmt.Errorf("RPC connection closed (service stopped, restarted, or wrong address): %w", err)
		}
		return err
	}
	return nil
}
