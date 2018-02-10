package parser

type Function struct {
	Name        string
	Arguments   []Field
	Results     []Field
	Imports     []string
	Package     string
	ServiceType string
}

//TODO cange name
type Field struct {
	Name    string
	Type    string
	Package string
}

type Type struct {
	Name   string
	Fields []Field
}
