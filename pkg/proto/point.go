package proto

// Point just provides line & column coordinates for a Result in a text file
type Point struct {
	Line   int `json:"line"   toml:"line"   yaml:"line"`
	Column int `json:"column" toml:"column" yaml:"column"`
}
