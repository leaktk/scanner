package scanner

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Request to the scanner to scan some resource
type Request struct {
	// Client provided identifier for associating a response to a request
	ID string `json:"id"`
	// Kind of thing being scanned
	Kind string `json:"kind"`
	// Thing to scan (e.g. URL, snippet of text, etc)
	Resource string `json:"resource"`
	// Flags to pass to the scanner (these depend heavily on the Kind)
	Options any `json:"options"`
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
		return err
	}

	r.ID = temp.ID
	r.Kind = temp.Kind
	r.Resource = temp.Resource

	// When adding different kinds, make sure to provide the corresponding
	// request.<Kind>Options() (*<Kind>) that returns nil if that isn't the
	// type of options on the object.
	switch temp.Kind {
	case "GitRepo":
		var options GitRepoOptions

		if err := json.Unmarshal(temp.Options, &options); err != nil {
			return err
		}

		r.Options = &options
	default:
		return fmt.Errorf("unsupported kind kind=%s id=%s", temp.Kind, temp.ID)
	}

	return nil
}

// GitRepoOptions returns Options casted to GitRepoOptions if that is the type
func (r *Request) GitRepoOptions() *GitRepoOptions {
	if options, ok := r.Options.(*GitRepoOptions); ok {
		return options
	}

	return nil
}

// GitRepoOptions stores options specific to GitRepo scan requests
type GitRepoOptions struct {
	// Only scan this many commits (reduced if larger than the max scan depth)
	Depth int `json:"depth"`
	// Only scan since this date
	Since string `json:"since"`
	// Only scan this branch
	Branch string `json:"branch"`
	// Work through a proxy for this request
	Proxy string `json:"proxy"`
}
