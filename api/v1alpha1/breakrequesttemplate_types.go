/*
Copyright 2025.

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
	"github.com/peak-scale/break-the-glass/api"
	"github.com/peak-scale/break-the-glass/internal/items"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BreakRequestTemplateSpec defines the desired state of BreakRequestTemplate.
type BreakRequestTemplateSpec struct {

	// Actual Items being created by this template
	// +kubebuilder:validation:Required
	Items items.TemplateItems `json:"items,omitempty"`

	// The default duration the BreakRequest referencing this template should be valid for.
	DefaultDuration metav1.Duration `json:"duration,omitempty"`
	// The duration this AccessRequest will be kept in the system after it has been expired (eg. auditing purposes)
	// If not set, the AccessRequest will be deleted after expiring.
	KeepFor api.ExtendedDuration `json:"keepFor,omitempty"`

	// AutoApprove requests created by this template will be automatically approved.
	AutoApprove bool `json:"autoApprove,omitempty"`

	// ApprovalCondition an optional CEL expression that must be successful for the request to be approved.
	ApprovalCondition string `json:"approvalCondition,omitempty"`
}

// BreakRequestTemplateStatus defines the observed state of BreakRequestTemplate.
type BreakRequestTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="AutoApprove",type=boolean,JSONPath=`.spec.autoApprove`
// +kubebuilder:printcolumn:name="Condition",type=string,JSONPath=`.spec.approvalCondition`,priority=10

// BreakRequestTemplate is the Schema for the breakrequesttemplates API.
type BreakRequestTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BreakRequestTemplateSpec   `json:"spec,omitempty"`
	Status BreakRequestTemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BreakRequestTemplateList contains a list of BreakRequestTemplate.
type BreakRequestTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BreakRequestTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BreakRequestTemplate{}, &BreakRequestTemplateList{})
}
