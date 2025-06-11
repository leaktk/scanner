package scanner

import (
	"context"

	"github.com/leaktk/leaktk/pkg/resource"
	"github.com/leaktk/leaktk/pkg/response"
)

// Backend is an interface for a scanner backend leveraged by leaktk
type Backend interface {
	Name() string
	Scan(ctx context.Context, resource resource.Resource) ([]*response.Result, error)
}
