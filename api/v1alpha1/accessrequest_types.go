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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AccessRequestSpec defines the desired state of AccessRequest.
type AccessRequestSpec struct {
	// AccessRuleReference name of the access rule to request
	// +kubebuilder:validation:Required
	Subject AccessRequestSubject `json:"subject"`
	// A reason on why the request is needed
	// +kubebuilder:validation:Required
	Reason string `json:"reason"`
	// Optional Time AccessRequests are valid for, after which they will be automatically deleted. Must be within max duration.
	// If not defined, defaults from the requested rule.
	Duration metav1.Duration `json:"duration"`
	// Optional point in time when the access should begin. Must be in the future.
	// If omitted, this is set to the current time. The Request must already be approved before the start time.
	// +optional
	// +kubebuilder:validation:Format=date-time
	// +kubebuilder:validation:Type=string
	StartTime *metav1.Time `json:"startTime,omitempty"`
}

// Relevant information about the subject of the access request (who wants what).
type AccessRequestSubject struct {
	Subjects []rbacv1.Subject `json:"subjects"`
	RoleRef  rbacv1.RoleRef   `json:"roleRef"`

	// Scope defines whether a RoleBinding oder ClusterRoleBinding is requested.
	// - Namespace: A RoleBinding is requested, which is limited to a specific namespace, where the AccessRequest was created.
	// - Cluster: A ClusterRoleBinding is requested, which grants access across the entire cluster.
	// +kubebuilder:validation:Enum=Cluster;Namespace
	// +kubebuilder:default=Namespace
	Scope SubjectScope `json:"scope,omitempty"`
}

type SubjectScope string

const (
	ScopeCluster   SubjectScope = "Cluster"
	ScopeNamespace SubjectScope = "Namespace"
)

// AccessRequestStatus defines the observed state of AccessRequest.
type AccessRequestStatus struct {
	// The requesting actor for the access request.
	// When the mutating webhook is enabled, this will be set to the user who created the request.
	Requestor AccessEntity `json:"requestor,omitempty"`
	// Reviewer refers to the subject that either approved or denied the request
	Reviewer AccessEntity `json:"reviewer,omitempty"`
	// The Approved propertoes are set when the request is approved.
	Approved AccessRequestStatusApproved `json:"approved,omitempty"`

	// Shows timestamps beetwen approval and termination of the request.
	Active AccessRequestStatusActive `json:"active,omitempty"`

	KeepUntil metav1.Timestamp `json:"keepUntil"`

	// conditions applied to the request. Known conditions are "Requested", "Denied", "Active", "Terminated" and "Failed". Latests condition is reflected in the phase.
	Conditions []AccessRequestStatusConditionItem `json:"conditions,omitempty"`

	// +kubebuilder:validation:Enum=Requested;Denied;Active;Terminated
	Phase RequestPhase `json:"phase,omitempty"`
}

// On Approval, we use all the properties from the spec and transform them into the status.
// This ensures that we are using the properties from that exact moment in time, when the request was approved (could be tampered with without admission).
type AccessRequestStatusApproved struct {
	Duration  metav1.Duration     `json:"duration"`
	StartTime *metav1.Time        `json:"startTime,omitempty"`
	Scope     SubjectScope        `json:"scope,omitempty"`
	Rules     []rbacv1.PolicyRule `json:"rules,omitempty"`
}

type AccessRequestStatusActive struct {
	ActiveFrom  metav1.Timestamp `json:"from"`
	ActiveUntil metav1.Timestamp `json:"until"`
}

type RequestPhase string

const (
	RequestPhasePending  RequestPhase = "Pending"
	RequestPhaseDenied   RequestPhase = "Denied"
	RequestPhaseApproved RequestPhase = "Approved"
	RequestPhaseActive   RequestPhase = "Active"
	RequestPhaseFailed   RequestPhase = "Failed"
	RequestPhaseExpired  RequestPhase = "Expired"
)

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
