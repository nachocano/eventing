/*
 * Copyright 2020 The Knative Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	Events             []Event           `json:"events,omitempty"`
}

type Event struct {
	Type              string                 `json:"type"`
	Description       string                 `json:"description,omitempty"`
	DataContentType   string                 `json:"datacontenttype,omitempty"`
	DataSchema        string                 `json:"dataschema,omitempty"`
	DataSchemaType    string                 `json:"dataschematype,omitempty"`
	DataSchemaContent string                 `json:"dataschemacontent,omitempty"`
	SourceTemplate    string                 `json:"sourcetemplate,omitempty"`
	Extensions        []eventingv1.Extension `json:"extensions,omitempty"`
}
