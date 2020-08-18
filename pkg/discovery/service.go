package discovery

type Service struct {
	Id                 string            `json:"id"`
	Name               string            `json:"name"`
	Url                string            `json:"url"`
	Description        string            `json:"description,omitempty"`
	DocsUrl            string            `json:"docsurl,omitempty"`
	SpecVersions       []string          `json:"specversions"`
	SubscriptionUrl    string            `json:"subscriptionurl"`
	SubscriptionConfig map[string]string `json:"subscriptionconfig,omitempty"`
	AuthScope          string            `json:"authscope,omitempty"`
	Protocols          []string          `json:"protocols"`
	Types              []Type            `json:"types,omitempty"`
}

type Type struct {
	Type              string      `json:"type"`
	Description       string      `json:"description,omitempty"`
	DataContentType   string      `json:"datacontenttype,omitempty"`
	DataSchema        string      `json:"dataschema,omitempty"`
	DataSchemaType    string      `json:"dataschematype,omitempty"`
	DataSchemaContent string      `json:"dataschemacontent,omitempty"`
	SourceTemplate    string      `json:"sourceTemplate,omitempty"`
	Extensions        []Extension `json:"extensions,omitempty"`
}

type Extension struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	SpecURL string `json:"specUrl,omitempty"`
}
