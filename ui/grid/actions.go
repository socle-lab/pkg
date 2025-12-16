package grid

import "errors"

type ActionScope string

const (
	ActionGlobal ActionScope = "global" // top bar / toolbar
	ActionRow    ActionScope = "row"    // per row
	ActionBulk   ActionScope = "bulk"   // on selected rows (optional)
)

type ActionIntent string

const (
	IntentCreate ActionIntent = "create"
	IntentView   ActionIntent = "view"
	IntentEdit   ActionIntent = "edit"
	IntentDelete ActionIntent = "delete"
	IntentExport ActionIntent = "export"
	IntentImport ActionIntent = "import"
	IntentCustom ActionIntent = "custom"
)

// ActionConfirm remains intention-only.
// Frontend decides how to render (modal, confirm(), toast, etc.)
type ActionConfirm struct {
	Title   string `json:"title,omitempty" yaml:"title,omitempty"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

type GridAction struct {
	ElementBase `json:",inline" yaml:",inline"`

	Scope  ActionScope  `json:"scope" yaml:"scope"`
	Intent ActionIntent `json:"intent" yaml:"intent"`

	// Target can be a route name, URL, or symbolic key (frontend decides)
	Target string `json:"target,omitempty" yaml:"target,omitempty"`

	// Optional: HTTP semantics are intent-ish, but still not “UI layout”
	Method string `json:"method,omitempty" yaml:"method,omitempty"` // GET/POST/DELETE/PATCH...

	Confirm *ActionConfirm `json:"confirm,omitempty" yaml:"confirm,omitempty"`

	// Optional authorization hook (backend intention)
	Permission string `json:"permission,omitempty" yaml:"permission,omitempty"`
}

func (a *GridAction) Kind() ElementKind { return KindAction }

func (a *GridAction) Normalize() {
	a.ElementBase.Normalize()
	if a.Scope == "" {
		a.Scope = ActionRow
	}
	if a.Intent == "" {
		a.Intent = IntentCustom
	}
	if a.Method == "" {
		// good default for “view/edit”
		a.Method = "GET"
	}
}

func (a GridAction) Validate() error {
	if err := a.ValidateBase(KindAction); err != nil {
		return err
	}
	switch a.Scope {
	case ActionGlobal, ActionRow, ActionBulk:
	default:
		return errors.New("action: invalid scope: " + string(a.Scope))
	}
	switch a.Intent {
	case IntentCreate, IntentView, IntentEdit, IntentDelete, IntentExport, IntentImport, IntentCustom:
	default:
		return errors.New("action: invalid intent: " + string(a.Intent))
	}
	return nil
}

// Actions container
type GridActions struct {
	Enabled bool         `json:"enabled" yaml:"enabled"`
	Items   []GridAction `json:"items" yaml:"items"`
}

func (ga *GridActions) Normalize() {
	for i := range ga.Items {
		ga.Items[i].Normalize()
	}
	//SortElements(ga.Items)
}

func (ga GridActions) Validate() error {
	for _, a := range ga.Items {
		if err := a.Validate(); err != nil {
			return err
		}
	}
	return nil
}
