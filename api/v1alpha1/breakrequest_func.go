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
	"context"
	"fmt"

	"github.com/peak-scale/break-the-glass/internal/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Sets Requests to pending
func (ar *BreakRequest) SetRequested() (err error) {
	if err := ar.transitionRequestPhase(
		RequestPhaseRequested,
		"Pending Review",
		"PendingReview",
		metav1.Now(),
		nil,
	); err != nil {
		return err
	}

	ar.Status.Review = &BreakRequestStatusReview{
		Verdict: RequestVerdictPending,
	}

	return
}

// Sets Requests to pending
func (ar *BreakRequest) SetPending() (err error) {
	if err := ar.transitionRequestPhase(
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
func (ar *BreakRequest) ApproveRequest(
	entity *AccessEntity,
	properties *BreakRequestStatusReviewProperties,
	reason string,
) (err error) {
	if reason == "" {
		reason = "Access request approved"
	}

	if err := ar.transitionRequestPhase(
		RequestPhaseApproved,
		reason,
		"ApprovedBy"+entity.Type.String(),
		metav1.Now(),
		entity,
	); err != nil {
		return err
	}

	ar.Status.Approved = properties

	ar.Status.Review = &BreakRequestStatusReview{
		Reviewer: entity,
		Verdict:  RequestVerdictApproved,
		Message:  reason,
	}

	return
}

// Denies the BreakRequest. It may directly transition to the Denied phase or set a reason for denial.
func (ar *BreakRequest) DenyRequest(entity *AccessEntity, reason string) (err error) {
	if reason == "" {
		reason = "Access request denied"
	}

	if err := ar.transitionRequestPhase(
		RequestPhaseDenied,
		reason,
		"DeniedByReviewer",
		metav1.Now(),
		entity,
	); err != nil {
		return err
	}

	ar.Status.Review = &BreakRequestStatusReview{
		Reviewer: entity,
		Verdict:  RequestVerdictDenied,
		Message:  reason,
	}

	return
}

// Activates the BreakRequest, allowing the subject to access the requested resources.
func (ar *BreakRequest) ActiveRequest(
	ctx context.Context,
	c client.Client,
	entity *AccessEntity,
) (err error) {
	now := metav1.Now()

	if err := ar.transitionRequestPhase(
		RequestPhaseActive,
		"Access request activated",
		"ActivatedBySystem",
		now,
		entity,
	); err != nil {
		return err
	}

	controllerutil.AddFinalizer(ar, meta.ControllerFinalizer)

	ar.Status.Active.ActiveFrom = now

	// If a duration was set, otherwise the lifecycle must be canceled  manually
	if ar.Spec.Duration.Duration > 0 {
		activeUntil := now.Add(ar.Spec.Duration.Duration)
		ar.Status.Active.ActiveUntil = metav1.NewTime(activeUntil)

		if ar.Spec.KeepFor.Duration > 0 {
			ar.Status.KeepUntil = metav1.NewTime(activeUntil.Add(ar.Spec.KeepFor.Duration))
		}

	}

	return nil
}

// When a request is active, it can be expired. This indicates that the granted access is revoked
// however this Request itself may be present longer, for auditing purposes
func (ar *BreakRequest) ExpireRequest(entity *AccessEntity) (err error) {
	if err := ar.transitionRequestPhase(
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
func (ar *BreakRequest) DeleteRequest(entity AccessEntity) {
	controllerutil.RemoveFinalizer(ar, meta.ControllerFinalizer)
}

// Get the Properties which are relevant for Review
func (ar *BreakRequest) GetReviewProperties(ctx context.Context, c client.Client) (*BreakRequestStatusReviewProperties, error) {
	return &BreakRequestStatusReviewProperties{
		Duration:  ar.Spec.Duration,
		StartTime: metav1.Now(),
		Items:     ar.Spec.Items,
	}, nil
}

// Ensure Phases are valid transitions and handle conditions accordingly
func (ar *BreakRequest) transitionRequestPhase(
	newPhase RequestPhase,
	conditionMessage string,
	reason string,
	now metav1.Time,
	entity *AccessEntity,
) error {

	// Prevent duplicate condition entries of the same type
	for _, cond := range ar.Status.Conditions {
		if RequestPhase(cond.Type) == newPhase {
			return nil
		}
	}

	// Disallow invalid transitions
	switch newPhase {
	case RequestPhaseDenied:
		if ar.Status.Phase == RequestPhaseApproved || ar.Status.Phase == RequestPhaseActive {
			return fmt.Errorf("cannot deny an already approved or active request")
		}
	case RequestPhaseApproved:
		if ar.Status.Phase == RequestPhaseDenied {
			return fmt.Errorf("cannot approve a denied request")
		}
	case RequestPhaseActive:
		if ar.Status.Phase != RequestPhaseApproved {
			return fmt.Errorf("can only activate an approved request")
		}
	case RequestPhaseExpired:
		if ar.Status.Phase != RequestPhaseActive {
			return fmt.Errorf("can only expire an active request")
		}
	}

	// Prevent duplicate condition entries of the same type
	for _, cond := range ar.Status.Conditions {
		if RequestPhase(cond.Type) == newPhase {
			return nil // Already in this state, no-op
		}
	}

	// Add new condition
	ar.Status.Conditions = append(
		[]metav1.Condition{{
			Type:               string(newPhase),
			Status:             metav1.ConditionTrue,
			Reason:             reason,
			Message:            conditionMessage,
			LastTransitionTime: now,
		}},
		ar.Status.Conditions...,
	)

	// Set the current phase
	ar.Status.Phase = newPhase

	return nil
}
