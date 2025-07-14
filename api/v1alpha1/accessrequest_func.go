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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
)

// Approves the AccessRequest. Depending on the start time, it may also directly activate the request.
func (ar *AccessRequest) ApproveRequest(entity AccessEntity, recorder record.EventRecorder) (err error) {
	if err := ar.transitionRequestPhase(
		RequestPhaseApproved,
		"Access request approved",
		"ApprovedByReviewer",
		metav1.Now(),
		&entity,
	); err != nil {
		return err
	}

	if recorder != nil {
		recorder.Event(ar, corev1.EventTypeNormal, "Approved", fmt.Sprintf("Request approved by %s %s", entity.Type, entity.Name))
	}

	return nil
}

// Denies the AccessRequest. It may directly transition to the Denied phase or set a reason for denial.
func (ar *AccessRequest) DenyRequest(entity AccessEntity, recorder record.EventRecorder, reason string) (err error) {
	if reason == "" {
		reason = "Access request denied"
	}

	if err := ar.transitionRequestPhase(
		RequestPhaseDenied,
		reason,
		"DeniedByReviewer",
		metav1.Now(),
		&entity,
	); err != nil {
		return err
	}

	if recorder != nil {
		recorder.Event(ar, corev1.EventTypeWarning, "Denied", fmt.Sprintf("Request denied by %s %s", entity.Type, entity.Name))
	}

	return nil
}

// Activates the AccessRequest, allowing the subject to access the requested resources.
func (ar *AccessRequest) ActiveRequest(recorder record.EventRecorder) (err error) {
	now := metav1.Now()

	if err := ar.transitionRequestPhase(
		RequestPhaseActive,
		"Access request activated",
		"ActivatedBySystem",
		now,
		nil,
	); err != nil {
		return err
	}

	ar.Status.Active.ActiveFrom = metav1.Timestamp{Seconds: now.Unix()}
	ar.Status.Active.ActiveUntil = metav1.Timestamp{Seconds: now.Add(ar.Spec.Duration.Duration).Unix()}

	if recorder != nil {
		recorder.Event(ar, corev1.EventTypeNormal, "Activated", fmt.Sprintf("Request Activated"))
	}

	return nil
}

func (ar *AccessRequest) ExpireRequest(entity AccessEntity, recorder record.EventRecorder) (err error) {
	if err := ar.transitionRequestPhase(
		RequestPhaseExpired,
		"Access request expired",
		"ExpiredBySystem",
		metav1.Now(),
		&entity,
	); err != nil {
		return err
	}

	if recorder != nil {
		recorder.Event(ar, corev1.EventTypeWarning, "Denied", fmt.Sprintf("Request Expired"))
	}

	return nil
}

// Ensure Phases are valid transitions and handle conditions accordingly
func (ar *AccessRequest) transitionRequestPhase(
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
	}

	// Prevent duplicate condition entries of the same type
	for _, cond := range ar.Status.Conditions {
		if RequestPhase(cond.Type) == newPhase {
			return nil // Already in this state, no-op
		}
	}

	// Add new condition
	ar.Status.Conditions = append(ar.Status.Conditions, AccessRequestStatusConditionItem{
		Type:               string(newPhase),
		Status:             corev1.ConditionTrue,
		Reason:             reason,
		Message:            conditionMessage,
		LastTransitionTime: now,
	})

	// Set reviewer (for terminal decisions)
	if entity != nil && (newPhase == RequestPhaseApproved || newPhase == RequestPhaseDenied) {
		ar.Status.Reviewer = *entity
	}

	// Set the current phase
	ar.Status.Phase = newPhase

	return nil
}
