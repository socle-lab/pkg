package grid

import "sort"

type FilterType string

const (
	FilterText    FilterType = "text"
	FilterSelect  FilterType = "select"
	FilterDate    FilterType = "date"
	FilterBoolean FilterType = "boolean"
)

type FilterOperator string

const (
	OpEquals   FilterOperator = "="
	OpContains FilterOperator = "contains"
	OpLike     FilterOperator = "like"
	OpGt       FilterOperator = ">"
	OpLt       FilterOperator = "<"
)

type HTMXFilter struct {
	Trigger string // change, keyup delay:500ms
	Target  string // #grid
	Swap    string // outerHTML
}

type FilterOption struct {
	Label string
	Value string
}

type GridFilter struct {
	Enabled       bool
	Fields        []GridFilterField
	FieldsPerLine int
}

type GridFilterField struct {
	Name        string     // lastname, email, enabled
	Label       string     // Nom, Email, Actif
	Type        FilterType // text, select, date, bool
	Operator    FilterOperator
	Position    int
	Enabled     bool
	Placeholder string
	Options     []FilterOption
	HTMX        *HTMXFilter
}

func (f *GridFilter) Sort() {
	sort.Slice(f.Fields, func(i, j int) bool {
		return f.Fields[i].Position < f.Fields[j].Position
	})
}

func (f *GridFilter) Normalize() {
	if f.FieldsPerLine <= 0 {
		f.FieldsPerLine = 3
	}
}
