package scanner

type (
	// Request to the scanner to scan some resource
	Request struct {
		ID       string            `json:"id"`
		Kind     string            `json:"kind"`
		Resource string            `json:"resource"`
		Options  map[string]string `json:"options"`
	}

	// Response from the scanner with the scan results
	Response struct {
		ID      string         `json:"id"`
		Request RequestDetails `json:"request"`
		Results []Result       `json:"results"`
	}

	// RequestDetails that we return with the response for tying the two together
	RequestDetails struct {
		ID       string `json:"id"`
		Kind     string `json:"kind"`
		Resource string `json:"resource"`
	}

	// Result of a scan
	Result struct {
		ID       string   `json:"id"`
		Kind     string   `json:"kind"`
		Match    string   `json:"match"`
		Context  string   `json:"context"`
		Entropy  float32  `json:"entropy"`
		Rule     Rule     `json:"rule"`
		Location Location `json:"location"`
		Contact  Contact  `json:"contact"`
		Date     string   `json:"date"`
		Note     string   `json:"note"`
	}

	// Location in the specific resource being scanned
	Location struct {
		Path string `json:"path"`
		// If the start column isn't available it will be zero.
		Start Point `json:"start"`
		// If the end information isn't available it will be the same as the
		// start information but the colmn will be the end of the line
		End Point `json:"end"`
	}

	// Point just provides line & column coordinates for a Result in a text file
	Point struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	}

	// Rule that triggered the result
	Rule struct {
		ID          string   `json:"id"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}

	// Contact for some resource when available
	Contact struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
)
