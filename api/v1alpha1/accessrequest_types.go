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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AccessRequestSpec defines the desired state of AccessRequest.
type AccessRequestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// AccessRuleReference name of the access rule to request
	// +kubebuilder:validation:Required
	AccessRuleReference string `json:"accessRuleReference"`

	// A reason on why the request is needed
	// +kubebuilder:validation:Required
	Reason string `json:"reason"`

	// Optional Time AccessRequests are valid for, after which they will be automatically deleted. Must be within max duration.
	// If not defined, defaults from the requested rule.
	Duration metav1.Duration `json:"duration"`

	// Optionally define for whom the request should be assigned. Defaults to the creation user of the request.
	For string `json:"for"`
}

// AccessRequestStatus defines the observed state of AccessRequest.
type AccessRequestStatus struct {
	ActiveFrom  metav1.Timestamp `json:"activeFrom"`
	ActiveUntil metav1.Timestamp `json:"activeUntil"`
	KeepUntil   metav1.Timestamp `json:"keepUntil"`
	Duration    metav1.Duration  `json:"duration"`

	// +kubebuilder:validation:Enum=Requested;Denied;Active;Terminated
	Phase RequestPhase `json:"phase"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Rule",type=string,JSONPath=`.spec.accessRuleReference`
// +kubebuilder:printcolumn:name="Reason",type=date,JSONPath=`.spec.reason`
// +kubebuilder:printcolumn:name="Reference",type=string,JSONPath=`.spec.reference`
// +kubebuilder:printcolumn:name="ActiveFrom",type=string,JSONPath=`.status.duration`,priority=10
// +kubebuilder:printcolumn:name="Duration",type=string,JSONPath=`.status.activeUntil`,priority=10
// +kubebuilder:printcolumn:name="ActiveUntil",type=string,JSONPath=`.status.activeFrom`,priority=10
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// AccessRequest is the Schema for the accessrequests API.
type AccessRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccessRequestSpec   `json:"spec,omitempty"`
	Status AccessRequestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AccessRequestList contains a list of AccessRequest.
type AccessRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AccessRequest{}, &AccessRequestList{})
}

type RequestPhase string

const (
	RequestPhaseRequested  RequestPhase = "Requested"
	RequestPhaseDenied     RequestPhase = "Denied"
	RequestPhaseActive     RequestPhase = "Active"
	RequestPhaseTerminated RequestPhase = "Terminated"
)
