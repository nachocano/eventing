package v1alpha1

import "knative.dev/pkg/apis"

type Schema struct {
	URI *apis.URL             `json:"uri,omitempty"`
	Ref *LocalSchemaReference `json:"ref,omitempty"`
}

type LocalSchemaReference struct {
	// TODO Schema group not needed
	SchemaGroup string `json:"schemaGroup,omitempty"`
	Name        string `json:"name"`
	Version     int32  `json:"version"`
}
