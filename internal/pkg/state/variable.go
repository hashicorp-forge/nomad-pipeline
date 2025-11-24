package state

type HCLVariable struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Default  any    `json:"default"`
	Required bool   `json:"required"`
}
