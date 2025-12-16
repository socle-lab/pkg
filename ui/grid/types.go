package grid

type Grid struct {
	Navbar GridNavbar
	Filter GridFilter
	Head   GridHead
}

// NAVBAR
type GridNavbar struct {
	Type    string
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

// FILTER
type GridFilter struct {
	BasicFilterKey         string
	AdvancedFilterFormPath string
}

// HEAD
type GridHead struct {
	Fields []GridField
}

// FIELD
type GridField struct {
	Name     string
	Label    string
	Sortable bool
	Position int
	Enabled  bool
}
