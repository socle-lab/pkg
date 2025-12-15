package grid

type Grid struct {
	Navbar GridNavbar
	Filter GridFilter
	Head   GridHead
}

// NAVBAR
type GridNavbar struct {
	Type    string
	Items   map[string]GridNavbarItem
	Enabled bool
}

type GridNavbarItem struct {
	Label   string
	Path    string
	Enabled bool
}

// FILTER
type GridFilter struct {
	BasicFilterKey         string
	AdvancedFilterFormPath string
}

// HEAD
type GridHead struct {
	Fields map[string]GridField
}

// FIELD
type GridField struct {
	Label    string
	Sortable bool
	Position int
	Enabled  bool
}
