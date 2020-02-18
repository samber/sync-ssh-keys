package datasources

type Source interface {
	GetName() string
	CheckInputErrors() string
	GetKeys() []string
}
