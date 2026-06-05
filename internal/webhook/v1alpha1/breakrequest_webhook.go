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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	addonsv1alpha1 "github.com/peak-scale/break-the-glass/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var brlog = logf.Log.WithName("br-resource")

// SetupBreakRequestWebhookWithManager registers the webhook for BreakRequest in the manager.
func SetupBreakRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &addonsv1alpha1.BreakRequest{}).
		WithValidator(&BreakRequestCustomValidator{client: mgr.GetClient()}).
		Complete()
}

// +kubebuilder:webhook:path=/validate-breakrequests,mutating=false,failurePolicy=fail,sideEffects=None,groups=addons.projectcapsule.dev,resources=breakrequests,verbs=create,versions=v1alpha1,name=vbr-v1alpha1.kb.io,admissionReviewVersions=v1

// BreakRequestCustomValidator struct is responsible for validating the BreakRequest resource
// when it is created, updated, or deleted.
type BreakRequestCustomValidator struct {
	client client.Client
}

var _ admission.Validator[*addonsv1alpha1.BreakRequest] = &BreakRequestCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type BreakRequest.
func (v *BreakRequestCustomValidator) ValidateCreate(
	_ context.Context,
	br *addonsv1alpha1.BreakRequest,
) (admission.Warnings, error) {
	brlog.Info("Validation for BreakRequest upon creation", "name", br.GetName())

	brt := &addonsv1alpha1.BreakRequestTemplate{}
	err := client.Reader(v.client).Get(
		context.Background(),
		client.ObjectKey{
			Name: br.Spec.TemplateName,
		},
		brt,
	)
	if err != nil {
		return nil, fmt.Errorf("error loading template %s: %w", br.Spec.TemplateName, err)
	}

	if brt.Spec.MaxDuration.Duration > 0 &&
		br.Spec.Duration.Duration > brt.Spec.MaxDuration.Duration {
		return nil, fmt.Errorf("requested duration %s exceeds template maxDuration %s",
			br.Spec.Duration.Duration, brt.Spec.MaxDuration.Duration)
	}

	_, err = br.RenderItemsItems(brt.Spec.Items)
	return nil, err
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type BreakRequest.
func (v *BreakRequestCustomValidator) ValidateUpdate(
	_ context.Context,
	oldBr, newBr *addonsv1alpha1.BreakRequest,
) (admission.Warnings, error) {
	if oldBr.Spec.TemplateName != newBr.Spec.TemplateName {
		return nil, fmt.Errorf(
			"templateName cannot be changed. old: %s, new: %s",
			oldBr.Spec.TemplateName,
			newBr.Spec.TemplateName,
		)
	}
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type BreakRequest.
func (v *BreakRequestCustomValidator) ValidateDelete(
	_ context.Context,
	_ *addonsv1alpha1.BreakRequest,
) (admission.Warnings, error) {
	return nil, nil
}
