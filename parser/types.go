package parser

type FunctionArgument struct {
	Name string
	Type string
}

type Function struct {
	Name      string
	Arguments []FunctionArgument
	Results   []FunctionArgument
	Imports   []string
	Package   string
}

type Field struct {
	Name string
	Type string
}

type Type struct {
	Name  string
	Field Field
}
