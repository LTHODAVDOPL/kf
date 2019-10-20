// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	serving "github.com/google/kf/third_party/knative-serving/pkg/apis/serving/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LongRunningProcess is a description of a 12-factor application encompassing
// routing, deployment, autoscaling, and revisions.
type LongRunningProcess struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec LongRunningProcesSpec `json:"spec,omitempty"`

	// +optional
	Status LongRunningProcessStatus `json:"status,omitempty"`
}

type LongRunningProcesSpec struct {
	// Template defines the App's runtime configuration.
	// +optional
	Template AppSpecTemplate `json:"template"`

	// Instances defines the scaling rules for the App.
	Instances AppSpecInstances `json:"instances,omitempty"`
}

type LongRunningProcessStatus struct {
	// Pull in the fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`

	// Fields brought in from the HorizontalPodAutoscaler

	// CurrentReplicas is the number of replicas for this LRP.
	CurrentReplicas int32 `json:"currentReplicas"`
	// DesiredReplicas is the desired number of replicas for this LRP.
	DesiredReplicas int32 `json:"desiredReplicas"`

	// Inline the latest serving.Service revisions that are ready
	serving.ConfigurationStatusFields `json:",inline"`

	// Inline the latest Service route information.
	serving.RouteStatusFields `json:",inline"`
}
