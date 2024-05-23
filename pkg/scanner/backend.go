package scanner

import (
	"github.com/leaktk/scanner/pkg/resource"
)

// Backend is an interface for a scanner backend leveraged by leaktk
type Backend interface {
	Scan(resource resource.Resource) ([]*Result, error)
}
