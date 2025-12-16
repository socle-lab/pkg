package grid

import "errors"

type FilterType string

const (
	FilterText    FilterType = "text"
	FilterSelect  FilterType = "select"
	FilterBoolean FilterType = "boolean"
	FilterDate    FilterType = "date"
	FilterNumber  FilterType = "number"
	FilterMulti   FilterType = "multi" // multi-select
)

type FilterOperator string

const (
	OpEquals   FilterOperator = "="
	OpContains FilterOperator = "contains"
	OpLike     FilterOperator = "like"
	OpGt       FilterOperator = ">"
	OpGte      FilterOperator = ">="
	OpLt       FilterOperator = "<"
	OpLte      FilterOperator = "<="
	OpIn       FilterOperator = "in"
)

type FilterOption struct {
	Label string `json:"label" yaml:"label"`
	Value string `json:"value" yaml:"value"`
}

type GridFilterField struct {
	ElementBase `json:",inline" yaml:",inline"`

	Type     FilterType     `json:"type" yaml:"type"`
	Operator FilterOperator `json:"operator" yaml:"operator"`

	Placeholder string         `json:"placeholder,omitempty" yaml:"placeholder,omitempty"`
	Options     []FilterOption `json:"options,omitempty" yaml:"options,omitempty"`
	// Optional default value (stringly typed on purpose; frontend decides input)
	Default string `json:"default,omitempty" yaml:"default,omitempty"`
}

func (f *GridFilterField) Kind() ElementKind { return KindFilter }

func (f *GridFilterField) Normalize() {
	f.ElementBase.Normalize()
	if f.Type == "" {
		f.Type = FilterText
	}
	if f.Operator == "" {
		// good default for most text inputs
		f.Operator = OpContains
	}
}

func (f GridFilterField) Validate() error {
	if err := f.ValidateBase(KindFilter); err != nil {
		return err
	}
	switch f.Type {
	case FilterText, FilterSelect, FilterBoolean, FilterDate, FilterNumber, FilterMulti:
	default:
		return errors.New("filter: invalid type: " + string(f.Type))
	}
	switch f.Operator {
	case OpEquals, OpContains, OpLike, OpGt, OpGte, OpLt, OpLte, OpIn:
	default:
		return errors.New("filter: invalid operator: " + string(f.Operator))
	}
	// If select/multi: options usually required
	if (f.Type == FilterSelect || f.Type == FilterMulti) && len(f.Options) == 0 {
		// not strictly required (could be remote), but helpful:
		// keep it as soft rule by placing a hint into Meta if you want.
	}
	return nil
}

// Filters container: intention only (fields per line, etc.)
type GridFilters struct {
	Enabled       bool              `json:"enabled" yaml:"enabled"`
	FieldsPerLine int               `json:"fields_per_line" yaml:"fields_per_line"`
	Fields        []GridFilterField `json:"fields" yaml:"fields"`
}

func (gf *GridFilters) Normalize() {
	if gf.FieldsPerLine <= 0 {
		gf.FieldsPerLine = 3
	}
	for i := range gf.Fields {
		gf.Fields[i].Normalize()
	}
	//SortElements(gf.Fields)
}

func (gf GridFilters) Validate() error {
	for _, f := range gf.Fields {
		if err := f.Validate(); err != nil {
			return err
		}
	}
	return nil
}
