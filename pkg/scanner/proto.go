package scanner

type (
	// Request to the scanner to scan some resource
	Request struct {
		Id       string            `json:"id"`
		Kind     string            `json:"kind"`
		Resource string            `json:"resource"`
		Options  map[string]string `json:"options"`
	}

	// Response from the scanner with the scan results
	Response struct {
		Id      string         `json:"id"`
		Request RequestDetails `json:"request"`
		Results []Result       `json:"results"`
	}

	// RequestDetails that we return with the response for tying the two together
	RequestDetails struct {
		Id       string `json:"id"`
		Kind     string `json:"kind"`
		Resource string `json:"resource"`
	}

	// Result of a scan
	Result struct {
		Id       string   `json:"id"`
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
		Path  string `json:"path"`
		Start Point  `json:"start"`
		End   Point  `json:"end"`
	}

	// Point just provides (x,y) coordinates for a Result in a text file
	Point struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	// Rule that triggered the result
	Rule struct {
		Id          string   `json:"id"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}

	// Contact for some resource
	Contact struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
)
