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

package subscription

// TODO move to apis

type Subscription struct {
	Id               string            `json:"id"`
	Protocol         string            `json:"protocol"`
	ProtocolSettings map[string]string `json:"protocolsettings,omitempty"`
	Sink             string            `json:"sink"`
	Filter           *FilterDialect    `json:"filter,omitempty"`
}

type FilterDialect struct {
	Dialect string   `json:"dialect"`
	Filters []Filter `json:"filters"`
}

type Filter struct {
	Type     string `json:"type"`
	Property string `json:"property"`
	Value    string `json:"value"`
}
