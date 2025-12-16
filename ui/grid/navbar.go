package grid

// Optional navbar (si tu veux le garder)
type GridNavbar struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
	// Intention-only: frontend decides rendering (tabs, pills, menu)
	Items []GridNavbarItem `json:"items" yaml:"items"`
}

type GridNavbarItem struct {
	ElementBase `json:",inline" yaml:",inline"`
	Path        string `json:"path" yaml:"path"`
}

func (n *GridNavbar) Normalize() {
	for i := range n.Items {
		n.Items[i].Normalize()
	}
	//SortElements(n.Items)
}

func (n GridNavbar) Validate() error {
	for _, it := range n.Items {
		if err := it.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (it *GridNavbarItem) Kind() ElementKind { return "navbar_item" }
func (it *GridNavbarItem) Normalize()        { it.ElementBase.Normalize() }
func (it GridNavbarItem) Validate() error    { return it.ValidateBase("navbar_item") }

func (g *Grid) Normalize() {
	g.Navbar.Normalize()
	g.Filter.Normalize()
	g.Head.Normalize()
	g.Actions.Normalize()
}

func (g Grid) Validate() error {
	if err := g.Navbar.Validate(); err != nil {
		return err
	}
	if err := g.Filter.Validate(); err != nil {
		return err
	}
	if err := g.Head.Validate(); err != nil {
		return err
	}
	if err := g.Actions.Validate(); err != nil {
		return err
	}
	return nil
}

func NewGridNavbarItem(
	name string,
	label string,
	path string,
) GridNavbarItem {
	item := GridNavbarItem{
		ElementBase: ElementBase{
			Name:     name,
			Label:    label,
			Visible:  true,
			Priority: 100,
		},
		Path: path,
	}

	item.Normalize()
	return item
}
