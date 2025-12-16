package grid

import "errors"

type FieldType string

const (
	FieldText    FieldType = "text"
	FieldEmail   FieldType = "email"
	FieldNumber  FieldType = "number"
	FieldBoolean FieldType = "boolean"
	FieldDate    FieldType = "date"
	FieldBadge   FieldType = "badge"
	FieldJSON    FieldType = "json"
)

type GridField struct {
	ElementBase `json:",inline" yaml:",inline"`

	Type     FieldType `json:"type" yaml:"type"`
	Sortable bool      `json:"sortable" yaml:"sortable"`
	// Optional: whether this column can be toggled by the user (frontend feature)
	Togglable bool `json:"togglable" yaml:"togglable"`
}

func (f *GridField) Kind() ElementKind { return KindField }

func (f *GridField) Normalize() {
	f.ElementBase.Normalize()
	if f.Type == "" {
		f.Type = FieldText
	}
	// sensible defaults:
	// Visible defaults handled by ElementBase.Normalize
}

func (f GridField) Validate() error {
	if err := f.ValidateBase(KindField); err != nil {
		return err
	}
	switch f.Type {
	case FieldText, FieldEmail, FieldNumber, FieldBoolean, FieldDate, FieldBadge, FieldJSON:
		return nil
	default:
		return errors.New("field: invalid type: " + string(f.Type))
	}
}

func NewGridField(
	name string,
	label string,
	fieldType FieldType,
	sortable bool,
	Type string,
	Togglable bool,
) GridField {
	f := GridField{
		ElementBase: ElementBase{
			Name:     name,
			Label:    label,
			Visible:  true,
			Priority: 100,
		},
		Type:      fieldType,
		Sortable:  false,
		Togglable: true,
	}

	f.Normalize()
	return f
}
