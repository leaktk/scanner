package scanner

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/leaktk/scanner/pkg/logger"
	"github.com/leaktk/scanner/pkg/resource"
)

// Request to the scanner to scan some resource
type Request struct {
	ID string
	// Thing to scan (e.g. URL, snippet of text, etc)
	Resource resource.Resource
	Errors   []LeakTKError
}

// UnmarshalJSON sets r to a copy of data
func (r *Request) UnmarshalJSON(data []byte) error {
	if r == nil {
		return errors.New("Request: UnmarshalJSON on nil pointer")
	}

	var temp struct {
		ID       string          `json:"id"`
		Kind     string          `json:"kind"`
		Resource string          `json:"resource"`
		Options  json.RawMessage `json:"options"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		logger.Debug("Request:\n%v", data)
		return fmt.Errorf("could not unmarshal Request: error=%q", err)
	}

	requestResource, err := resource.NewResource(temp.Kind, temp.Resource, temp.Options)
	if err != nil {
		return fmt.Errorf("could not create resource: error=%q", err)
	}

	r.ID = temp.ID
	r.Resource = requestResource

	return nil
}
