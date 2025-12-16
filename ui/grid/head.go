package grid

type GridHead struct {
	Fields []GridField `json:"fields" yaml:"fields"`
}

func (h *GridHead) Normalize() {
	for i := range h.Fields {
		h.Fields[i].Normalize()
	}
	//SortElements(h.Fields)
}

func (h GridHead) Validate() error {
	for _, f := range h.Fields {
		if err := f.Validate(); err != nil {
			return err
		}
	}
	return nil
}
