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
	"fmt"
	"time"

	"github.com/peak-scale/break-the-glass/internal/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Sets Requests to pending
func (br *BreakRequest) SetRequested() (err error) {
	if err := br.transitionRequestPhase(
		RequestPhaseRequested,
		"Pending Review",
		"PendingReview",
		metav1.Now(),
		nil,
	); err != nil {
		return err
	}

	br.Status.Review = &BreakRequestStatusReview{
		Verdict: RequestVerdictPending,
	}

	return
}

// Sets Requests to pending
func (br *BreakRequest) SetPending() (err error) {
	if err := br.transitionRequestPhase(
		RequestPhasePending,
		"Access request pending",
		"PendingBySystem",
		metav1.Now(),
		nil,
	); err != nil {
		return err
	}

	return
}

// Approves the BreakRequest. Depending on the start time, it may also directly activate the request.
func (br *BreakRequest) ApproveRequest(
	entity *AccessEntity,
	properties *BreakRequestStatusReviewProperties,
	reason string,
) (err error) {
	if reason == "" {
		reason = "Access request approved"
	}

	if err := br.transitionRequestPhase(
		RequestPhaseApproved,
		reason,
		"ApprovedBy"+entity.Type.String(),
		metav1.Now(),
		entity,
	); err != nil {
		return err
	}

	// items are set by the controller, remove them from the status
	properties.Items = nil

	br.Status.Approved = properties

	br.Status.Review = &BreakRequestStatusReview{
		Reviewer: entity,
		Verdict:  RequestVerdictApproved,
		Message:  reason,
	}

	return err
}

// Denies the BreakRequest. It may directly transition to the Denied phase or set a reason for denial.
func (br *BreakRequest) DenyRequest(entity *AccessEntity, reason string) (err error) {
	if reason == "" {
		reason = "Access request denied"
	}

	if err := br.transitionRequestPhase(
		RequestPhaseDenied,
		reason,
		"DeniedByReviewer",
		metav1.Now(),
		entity,
	); err != nil {
		return err
	}

	br.Status.Review = &BreakRequestStatusReview{
		Reviewer: entity,
		Verdict:  RequestVerdictDenied,
		Message:  reason,
	}

	return
}

// Activates the BreakRequest, allowing the subject to access the requested resources.
func (br *BreakRequest) ActiveRequest(
	entity *AccessEntity,
) (err error) {
	now := metav1.Now()

	if err := br.transitionRequestPhase(
		RequestPhaseActive,
		"Access request activated",
		"ActivatedBySystem",
		now,
		entity,
	); err != nil {
		return err
	}

	controllerutil.AddFinalizer(br, meta.ControllerFinalizer)

	if br.Status.Active == nil {
		br.Status.Active = &BreakRequestStatusActive{}
	}

	br.Status.Active.ActiveFrom = now

	// If a duration was set, otherwise the lifecycle must be canceled manually
	if br.Spec.Duration.Duration > 0 {
		activeUntil := now.Add(br.Spec.Duration.Duration)
		br.Status.Active.ActiveUntil = metav1.NewTime(activeUntil)

		if br.Spec.KeepFor > 0 {
			br.Status.KeepUntil = metav1.NewTime(activeUntil.Add(time.Duration(br.Spec.KeepFor)))
		}
	}

	return nil
}

// When a request is active, it can be expired. This indicates that the granted access is revoked
// however this Request itself may be present longer, for auditing purposes
func (br *BreakRequest) ExpireRequest(entity *AccessEntity) (err error) {
	if err := br.transitionRequestPhase(
		RequestPhaseExpired,
		"Access request expired",
		"ExpiredBySystem",
		metav1.Now(),
		entity,
	); err != nil {
		return err
	}

	return
}

// Final stage, delete the request
func (br *BreakRequest) DeleteRequest() {
	controllerutil.RemoveFinalizer(br, meta.ControllerFinalizer)
}

// Get the Properties which are relevant for Review
func (br *BreakRequest) GetReviewProperties() (*BreakRequestStatusReviewProperties, error) {
	return &BreakRequestStatusReviewProperties{
		Duration:  br.Spec.Duration,
		StartTime: metav1.Now(),
		Items:     br.Spec.Items,
		KeepFor:   br.Spec.KeepFor,
	}, nil
}

// Ensure Phases are valid transitions and handle conditions accordingly
func (br *BreakRequest) transitionRequestPhase(
	newPhase RequestPhase,
	conditionMessage string,
	reason string,
	now metav1.Time,
	entity *AccessEntity,
) error {

	// Prevent duplicate condition entries of the same type
	for _, cond := range br.Status.Conditions {
		if RequestPhase(cond.Type) == newPhase {
			return nil
		}
	}

	// Disallow invalid transitions
	switch newPhase {
	case RequestPhaseDenied:
		if br.Status.Phase == RequestPhaseApproved || br.Status.Phase == RequestPhaseActive {
			return fmt.Errorf("cannot deny an already approved or active request")
		}
		setReviewer(br, entity, conditionMessage, RequestVerdictDenied)

	case RequestPhaseApproved:
		if br.Status.Phase == RequestPhaseDenied {
			return fmt.Errorf("cannot approve a denied request")
		}
		setReviewer(br, entity, conditionMessage, RequestVerdictApproved)

	case RequestPhaseActive:
		if br.Status.Phase != RequestPhaseApproved {
			return fmt.Errorf("can only activate an approved request")
		}

	case RequestPhaseExpired:
		if br.Status.Phase != RequestPhaseActive {
			return fmt.Errorf("can only expire an active request")
		}
	}

	// Prevent duplicate condition entries of the same type
	for _, cond := range br.Status.Conditions {
		if RequestPhase(cond.Type) == newPhase {
			return nil // Already in this state, no-op
		}
	}

	// Add new condition
	br.Status.Conditions = append(
		[]metav1.Condition{{
			Type:               string(newPhase),
			Status:             metav1.ConditionTrue,
			Reason:             reason,
			Message:            conditionMessage,
			LastTransitionTime: now,
		}},
		br.Status.Conditions...,
	)

	// Set the current phase
	br.Status.Phase = newPhase

	return nil
}

func setReviewer(
	ar *BreakRequest,
	entity *AccessEntity,
	conditionMessage string,
	verdict RequestVerdict,
) {
	if entity != nil {
		ar.Status.Review = &BreakRequestStatusReview{
			Reviewer: entity,
			Message:  conditionMessage,
			Verdict:  verdict,
		}
	}
}
