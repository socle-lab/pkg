package handler

import (
	socle "github.com/socle-lab/core"
)

// Handlers is the type for handlers, and gives access to Socle and models
type Handler struct {
	Core *socle.Socle
}
