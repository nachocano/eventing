/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/google/go-cmp/cmp"
	eventingduckv1alpha1 "knative.dev/eventing/pkg/apis/duck/v1alpha1"
)

func TestParallelSetDefaults(t *testing.T) {
	testCases := map[string]struct {
		nilChannelDefaulter bool
		channelTemplate     *eventingduckv1alpha1.ChannelTemplateSpec
		initial             Parallel
		expected            Parallel
	}{
		"nil ChannelDefaulter": {
			nilChannelDefaulter: true,
			expected:            Parallel{},
		},
		"unset ChannelDefaulter": {
			expected: Parallel{},
		},
		"set ChannelDefaulter": {
			channelTemplate: defaultChannelTemplate,
			expected: Parallel{
				Spec: ParallelSpec{
					ChannelTemplate: defaultChannelTemplate,
				},
			},
		},
		"template already specified": {
			channelTemplate: defaultChannelTemplate,
			initial: Parallel{
				Spec: ParallelSpec{
					ChannelTemplate: &eventingduckv1alpha1.ChannelTemplateSpec{
						TypeMeta: v1.TypeMeta{
							APIVersion: SchemeGroupVersion.String(),
							Kind:       "OtherChannel",
						},
					},
				},
			},
			expected: Parallel{
				Spec: ParallelSpec{
					ChannelTemplate: &eventingduckv1alpha1.ChannelTemplateSpec{
						TypeMeta: v1.TypeMeta{
							APIVersion: SchemeGroupVersion.String(),
							Kind:       "OtherChannel",
						},
					},
				},
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			if !tc.nilChannelDefaulter {
				eventingduckv1alpha1.ChannelDefaulterSingleton = &parallelChannelDefaulter{
					channelTemplate: tc.channelTemplate,
				}
				defer func() { eventingduckv1alpha1.ChannelDefaulterSingleton = nil }()
			}
			tc.initial.SetDefaults(context.TODO())
			if diff := cmp.Diff(tc.expected, tc.initial); diff != "" {
				t.Fatalf("Unexpected defaults (-want, +got): %s", diff)
			}
		})
	}
}

type parallelChannelDefaulter struct {
	channelTemplate *eventingduckv1alpha1.ChannelTemplateSpec
}

func (cd *parallelChannelDefaulter) GetDefault(_ string) *eventingduckv1alpha1.ChannelTemplateSpec {
	return cd.channelTemplate
}
