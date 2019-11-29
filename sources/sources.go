package sources

type Source interface {
	GetName() string
	CheckInputErrors() string
	GetKeys() []string
}
