package grid

import "sort"

func (h *GridHead) SortFields() {
	sort.Slice(h.Fields, func(i, j int) bool {
		return h.Fields[i].Position < h.Fields[j].Position
	})
}

func (n *GridNavbar) SortItems() {
	sort.Slice(n.Items, func(i, j int) bool {
		return n.Items[i].Position < n.Items[j].Position
	})
}
