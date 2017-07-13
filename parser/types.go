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
