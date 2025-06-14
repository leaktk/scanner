package proto

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/leaktk/leaktk/pkg/logger"
)

type RequestKind int

const (
	ContainerImageRequestKind RequestKind = itoa
	FilesRequestKind          RequestKind
	GitRepoRequestKind        RequestKind
	JSONDataRequestKind       RequestKind
	TextRequestKind           RequestKind
	URLRequestKind            RequestKind
)

type Request struct {
	ID       string
	Kind     RequestKind
	Resource string
	Options  any
}

// UnmarshalJSON sets r to a copy of data
func (r *Request) UnmarshalJSON(data []byte) error {
	if r == nil {
		return errors.New("Request: UnmarshalJSON on nil pointer")
	}

	var tmp struct {
		ID       string          `json:"id"`
		Kind     string          `json:"kind"`
		Resource string          `json:"resource"`
		Options  json.RawMessage `json:"options"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		logger.Debug("Request:\n%v", data)
		return fmt.Errorf("could not unmarshal request: %w", err)
	}

	r.ID = tmp.ID
	r.Resource = tmp.Resource

	switch tmp.Kind {
	case "GitRepo":
		r.Kind = GitRepoRequestKind
		r.Options = &GitRepoOptions{}
	case "JSONData":
		r.Kind = JSONDataRequestKind
		r.Options = &JSONDataOptions{}
	case "Files":
		r.Kind = FilesRequestKind
		r.Options = &CommonOptions{}
	case "Text":
		r.Kind = TextRequestKind
		r.Options = &CommonOptions{}
	case "URL":
		r.Kind = URLRequestKind
		r.Options = &CommonOptions{}
	case "ContainerImage":
		r.Kind = ContainerImageRequestKind
		r.Options = &ContainerImageOptions{}
	default:
		return fmt.Errorf("unsupported request kind: kind=%q", tmp.Kind)
	}

	if len(tmp.Options) > 0 {
		if err := json.Unmarshal(tmp.Options, kind.Options); err != nil {
			logger.Debug("%s:\n%v", tmp.Kind, tmp.Options)
			return nil, fmt.Errorf("could not unmarshal %s: %w", tmp.Kind, err)
		}
	}

	return nil
}
