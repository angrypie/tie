package parser

type Function struct {
	Name        string
	Arguments   []Field
	Results     []Field
	Receiver    Field
	Package     string
	ServiceType string
}

type Field struct {
	Name    string
	Type    string
	Package string
	Prefix  string
}

type Type struct {
	Name   string
	Fields []Field
}
