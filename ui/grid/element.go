package grid

import (
	"errors"
	"sort"
	"strings"
)

//
// Common: GridElement philosophy
//

type ElementKind string

const (
	KindField  ElementKind = "field"
	KindFilter ElementKind = "filter"
	KindAction ElementKind = "action"
)

type GridElement interface {
	Kind() ElementKind
	GetName() string
	GetLabel() string
	IsVisible() bool
	GetPriority() int
	Validate() error
}

// ElementBase carries the shared “intention” properties.
// - Name: stable key (used by frontend, query builders, templates)
// - Label: human label
// - Visible: should be displayed by default
// - Priority: ordering intent (1 = higher priority)
// - Meta: extra UI-agnostic hints (frontend can read them)
type ElementBase struct {
	Name     string            `json:"name" yaml:"name"`
	Label    string            `json:"label" yaml:"label"`
	Visible  bool              `json:"visible" yaml:"visible"`
	Priority int               `json:"priority" yaml:"priority"`
	Meta     map[string]string `json:"meta,omitempty" yaml:"meta,omitempty"`
}

func (b ElementBase) GetName() string  { return b.Name }
func (b ElementBase) GetLabel() string { return b.Label }
func (b ElementBase) IsVisible() bool  { return b.Visible }
func (b ElementBase) GetPriority() int { return b.Priority }
func (b *ElementBase) Normalize() {
	b.Name = strings.TrimSpace(b.Name)
	b.Label = strings.TrimSpace(b.Label)
	if b.Priority == 0 {
		// sensible default: 100 means “normal”; set 1..n for important elements
		b.Priority = 100
	}
	// default visible = true (unless explicitly set false)
	// If you prefer explicit, remove this block.
	if b.Meta == nil {
		b.Meta = map[string]string{}
	}
}

func (b ElementBase) ValidateBase(kind ElementKind) error {
	if b.Name == "" {
		return errors.New(string(kind) + ": name is required")
	}
	// Label can be empty if frontend wants to infer it from Name.
	return nil
}

//
// Sorting helpers (generic)
//

// SortElements sorts any GridElement slice by priority asc, then name asc.
// Lower priority value = more important / earlier.
func SortElements[T GridElement](items []T) {
	sort.SliceStable(items, func(i, j int) bool {
		pi, pj := items[i].GetPriority(), items[j].GetPriority()
		if pi != pj {
			return pi < pj
		}
		return items[i].GetName() < items[j].GetName()
	})
}
