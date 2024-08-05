package tui

type RoleItem struct {
	id      string
	name    string
	persona string
}

// implement the list.Item interface
func (i RoleItem) FilterValue() string {
	return i.persona
}

// implement the list.Item interface
func (i RoleItem) Title() string {
	return i.name
}

// implement the list.Item interface
func (i RoleItem) Description() string {
	return ""
}

func (i RoleItem) ID() string {
	return i.id
}

func (i RoleItem) Name() string {
	return i.name
}

func (i RoleItem) Persona() string {
	return i.persona
}
