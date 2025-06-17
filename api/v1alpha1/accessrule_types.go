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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AccessRuleSpec defines the desired state of AccessRule.
type AccessRuleSpec struct {
	// Time AccessRequests are valid for, after which they will be automatically deleted.
	// +kubebuilder:default="60m"
	DefaultDuration metav1.Duration `json:"duration"`

	// Time AccessRequests are valid for, after which they will be automatically deleted.
	// +kubebuilder:default="6h"
	MaximumDuration metav1.Duration `json:"maxDuration"`

	// The Duration for which AccessRequests will be kept in the system after they are expired (auditing reasons).
	// +kubebuilder:default="6h"
	KeepDuration metav1.Duration `json:"keepDuration"`

	// Subjects is the context from where AccessRequests (or from whom) can be created and assigned to this Rule.
	// +kubebuilder:validation:Required
	Subjects []string `json:"subjects"`

	AllowedRoles []AccessRuleAllowedRole `json:"allowedRoles"`

	// Action to execute when the AccessRule is matched.
	// +kubebuilder:default=AutoApprove
	// +kubebuilder:validation:Enum=AutoApprove;Deny
	Action string `json:"action,omitempty"`
}

// AccessRuleSpec defines the desired state of AccessRule.
type AccessRuleAllowedRole struct {
	AllowedRoles []string `json:"allowedRoles"`

	// Whether this rule applies to the namespace scope of the Access Request or the cluster scope.
	// +kubebuilder:default=true
	Namespaced bool `json:"namespaced,omitempty"`
}

// AccessRuleStatus defines the observed state of AccessRule.
type AccessRuleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// AccessRule is the Schema for the accessrules API.
type AccessRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccessRuleSpec   `json:"spec,omitempty"`
	Status AccessRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AccessRuleList contains a list of AccessRule.
type AccessRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AccessRule{}, &AccessRuleList{})
}
