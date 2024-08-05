package tui

type SkillItem struct {
	id          string
	instruction string
}

// implement the list.Item interface
func (i SkillItem) Title() string {
	return i.id
}

// implement the list.Item interface
func (i SkillItem) Description() string {
	return ""
}

// implement the list.Item interface
func (i SkillItem) FilterValue() string {
	return i.instruction
}

func (i SkillItem) ID() string {
	return i.id
}

func (i SkillItem) Instruction() string {
	return i.instruction
}
