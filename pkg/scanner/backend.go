package scanner

import (
	"github.com/leaktk/leaktk/pkg/resource"
	"github.com/leaktk/leaktk/pkg/response"
)

// Backend is an interface for a scanner backend leveraged by leaktk
type Backend interface {
	Name() string
	Scan(resource resource.Resource) ([]*response.Result, error)
}
