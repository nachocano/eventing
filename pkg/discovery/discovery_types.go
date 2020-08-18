package discovery

// TODO move to apis

import (
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
)

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
	Type              string                 `json:"type"`
	Description       string                 `json:"description,omitempty"`
	DataContentType   string                 `json:"datacontenttype,omitempty"`
	DataSchema        string                 `json:"dataschema,omitempty"`
	DataSchemaType    string                 `json:"dataschematype,omitempty"`
	DataSchemaContent string                 `json:"dataschemacontent,omitempty"`
	SourceTemplate    string                 `json:"sourceTemplate,omitempty"`
	Extensions        []eventingv1.Extension `json:"extensions,omitempty"`
}
