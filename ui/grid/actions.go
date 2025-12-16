package grid

import "sort"

type GridActions struct {
	Global []GridAction
	Row    []GridAction
}

type GridAction struct {
	Name     string
	Label    string
	Path     string
	Method   string // GET, POST, DELETE, PATCH
	Icon     string
	Enabled  bool
	Position int

	Confirm    *ActionConfirm
	Permission string
	HTMX       *HTMXAction
}

type ActionConfirm struct {
	Title   string
	Message string
}

type HTMXAction struct {
	Method  string // hx-get, hx-post, hx-delete
	Target  string // #modal, #content
	Swap    string // innerHTML, outerHTML
	Confirm bool   // hx-confirm
}

func (a *GridActions) Sort() {
	sort.Slice(a.Global, func(i, j int) bool {
		return a.Global[i].Position < a.Global[j].Position
	})
	sort.Slice(a.Row, func(i, j int) bool {
		return a.Row[i].Position < a.Row[j].Position
	})
}
