package proto

type CommonOptions struct {
	Priority int `json:"priority"`
}

type ContainerImageOptions struct {
	CommonOptions
	Arch       string   `json:"arch"`
	Depth      int      `json:"depth"`
	Exclusions []string `json:"exclusions"`
	Since      string   `json:"since"`
}

type GitRepoOptions struct {
	CommonOptions
	Branch   string `json:"branch"`
	Depth    int    `json:"depth"`
	Local    bool   `json:"local"`
	Staged   bool   `json:"staged"`
	Since    string `json:"since"`
	Proxy    string `json:"proxy"`
	Unstaged bool   `json:"unstaged"`
}

type JSONDataOptions struct {
	CommonOptions
	FetchURLs string `json:"fetch_urls"`
}
