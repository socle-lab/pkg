package grid

import "sort"

type GridNavbar struct {
	Items   []GridNavbarItem
	Enabled bool
}

type GridNavbarItem struct {
	Name     string
	Label    string
	Path     string
	Enabled  bool
	Position int
}

func (n *GridNavbar) SortItems() {
	sort.Slice(n.Items, func(i, j int) bool {
		return n.Items[i].Position < n.Items[j].Position
	})
}
