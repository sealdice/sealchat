package protocol

type Element struct {
	KElement bool      `json:"kElement"`
	Type     string    `json:"type"`
	Attrs    Dict      `json:"attrs"`
	Data     Dict      `json:"data"` // Deprecated
	Children []Element `json:"children"`
	Source   string    `json:"source"`
}

type Dict map[string]interface{}
