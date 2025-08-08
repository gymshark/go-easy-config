package config

type Loader[T any] interface {
	Load(c *T) error
}
