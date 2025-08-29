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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// BreakRequestSpec defines the desired state of BreakRequest.
type BreakRequestSpec struct {

	// TemplateName the name of the template to use for this request
	// +kubebuilder:validation:Required
	TemplateName string `json:"templateName"`
	// Requesting actor for the access request.
	Requestor AccessEntity `json:"requestor,omitempty"`
	// Actual Items being requested
	// +kubebuilder:validation:Required
	Items []runtime.RawExtension `json:"items,omitempty"`
	// A reason on why the request is needed
	Reason string `json:"reason,omitempty"`
	// The duration this BreakRequest should be valid for.
	// If no duration was defined the lifecycle is bound to the request itself -
	// if the request is deleted, it's the end of the duration.
	// The Request can also be Terminated by another automation via calling the ExpireRequest() API-Function.
	Duration metav1.Duration `json:"duration,omitempty"`
	// The duration this BreakRequest will be kept in the system after it has been expired (eg. auditing purposes)
	// If not set, the BreakRequest will be deleted after expiring.
	KeepFor api.ExtendedDuration `json:"keepFor,omitempty"`
	// Optional point in time when the access should begin. Must be in the future.
	// If omitted, this is set to the current time. The Request must already be approved before the start time.
	// +optional
	// +kubebuilder:validation:Format=date-time
	// +kubebuilder:validation:Type=string
	StartTime *metav1.Time `json:"startTime,omitempty"`
}

type SubjectScope string

const (
	ScopeCluster   SubjectScope = "Cluster"
	ScopeNamespace SubjectScope = "Namespace"
)

// BreakRequestStatus defines the observed state of BreakRequest.
type BreakRequestStatus struct {
	// Reviewer refers to the subject that either approved or denied the request
	Review *BreakRequestStatusReview `json:"review,omitempty"`
	// The Approved properties are set when the request is approved.
	Approved *BreakRequestStatusReviewProperties `json:"approved,omitempty"`
	// Shows timestamps beetwen approval and termination of the request.
	Active *BreakRequestStatusActive `json:"active,omitempty"`
	// The time when the request was created.
	KeepUntil metav1.Time `json:"keepUntil,omitempty"`
	// conditions applied to the request.
	// Known conditions are "Requested", "Pending", "Denied", "Approved", "Active" and "Expired".
	// Latests condition is reflected in the phase.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// +kubebuilder:validation:Enum=Requested;Pending;Denied;Approved;Active;Expired
	Phase RequestPhase `json:"phase,omitempty"`
}

type BreakRequestStatusActive struct {
	ActiveFrom  metav1.Time `json:"from,omitempty"`
	ActiveUntil metav1.Time `json:"until,omitempty"`
}

// These are the relevant properties which are subject to review and then persistet
type BreakRequestStatusReviewProperties struct {
	KeepFor   api.ExtendedDuration   `json:"keepFor,omitempty"`
	Duration  metav1.Duration        `json:"duration,omitempty"`
	StartTime metav1.Time            `json:"startTime,omitempty"`
	Items     []runtime.RawExtension `json:"items,omitempty"`
}

type BreakRequestStatusReview struct {
	// The Entity revieweing this request
	Reviewer *AccessEntity `json:"reviewer,omitempty"`
	// The verdict made by the reviewing entity
	// +kubebuilder:validation:Enum=Pending;Denied;Approved
	Verdict RequestVerdict `json:"verdict,omitempty"`
	// Message with the review
	Message string `json:"message,omitempty"`
}

type RequestVerdict string

const (
	RequestVerdictDenied   RequestVerdict = "Denied"
	RequestVerdictApproved RequestVerdict = "Approved"
	RequestVerdictPending  RequestVerdict = "Pending"
)

type RequestPhase string

const (
	RequestPhaseRequested RequestPhase = "Requested"
	RequestPhasePending   RequestPhase = "Pending"
	RequestPhaseDenied    RequestPhase = "Denied"
	RequestPhaseApproved  RequestPhase = "Approved"
	RequestPhaseActive    RequestPhase = "Active"
	RequestPhaseExpired   RequestPhase = "Expired"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Reason",type=date,JSONPath=`.spec.reason`
// +kubebuilder:printcolumn:name="Verdict",type=string,JSONPath=`.status.review.verdict`
// +kubebuilder:printcolumn:name="ActiveFrom",type=string,JSONPath=`.status.duration`,priority=10
// +kubebuilder:printcolumn:name="Duration",type=string,JSONPath=`.status.activeUntil`,priority=10
// +kubebuilder:printcolumn:name="ActiveUntil",type=string,JSONPath=`.status.activeFrom`,priority=10
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// BreakRequest is the Schema for the BreakRequests API.
type BreakRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BreakRequestSpec   `json:"spec,omitempty"`
	Status BreakRequestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BreakRequestList contains a list of BreakRequest.
type BreakRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BreakRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BreakRequest{}, &BreakRequestList{})
}
