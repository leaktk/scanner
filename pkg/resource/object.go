package resource

import "io"

// Object is a scanable yielded by the Objects function on a resource
type Object struct {
	// Path is the relative path to the object
	Path string
	// Content is a reader for accessing the object's content. Different objects
	// may provide different reader types, but assume this can only be read from
	// once
	Content io.Reader
}
