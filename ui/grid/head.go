package grid

import "sort"

type GridHead struct {
	Fields []GridField
}

type GridField struct {
	Name     string
	Label    string
	Sortable bool
	Position int
	Enabled  bool
}

func (h *GridHead) SortFields() {
	sort.Slice(h.Fields, func(i, j int) bool {
		return h.Fields[i].Position < h.Fields[j].Position
	})
}
