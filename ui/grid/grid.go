package grid

type Grid struct {
	Navbar  GridNavbar  `json:"navbar" yaml:"navbar"`
	Filter  GridFilter  `json:"filter" yaml:"filter"`
	Head    GridHead    `json:"head" yaml:"head"`
	Actions GridActions `json:"actions" yaml:"actions"`
}
