package aws

import (
	"github.com/ianlopshire/go-ssm-config"
)

type SSMParameterStoreLoader[T any] struct {
	Path string
}

func (s *SSMParameterStoreLoader[T]) Load(c *T) error {
	return ssmconfig.Process(s.Path, c)
}
