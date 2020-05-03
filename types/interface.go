package types

type Field interface {
	Name() string
	TypeName() string
	TypeParts() (string, string, string)
}
