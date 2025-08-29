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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	addonsv1alpha1 "github.com/peak-scale/break-the-glass/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var brtlog = logf.Log.WithName("brt-resource")

// SetupBreakRequestTemplateWebhookWithManager registers the webhook for BreakRequestTemplate in the manager.
func SetupBreakRequestTemplateWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&addonsv1alpha1.BreakRequestTemplate{}).
		WithValidator(&BreakRequestTemplateCustomValidator{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-addons-projectcapsule-dev-v1alpha1-brt,mutating=false,failurePolicy=fail,sideEffects=None,groups=addons.projectcapsule.dev,resources=brts,verbs=create;update,versions=v1alpha1,name=vbrt-v1alpha1.kb.io,admissionReviewVersions=v1

// BreakRequestTemplateCustomValidator struct is responsible for validating the BreakRequestTemplate resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type BreakRequestTemplateCustomValidator struct{}

var _ webhook.CustomValidator = &BreakRequestTemplateCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type BreakRequestTemplate.
func (v *BreakRequestTemplateCustomValidator) ValidateCreate(
	_ context.Context,
	obj runtime.Object,
) (admission.Warnings, error) {
	brt, ok := obj.(*addonsv1alpha1.BreakRequestTemplate)
	if !ok {
		return nil, fmt.Errorf("expected a BreakRequestTemplate object but got %T", obj)
	}
	brtlog.Info("Validation for BreakRequestTemplate upon creation", "name", brt.GetName())

	return nil, validate(brt)
}
func validate(brt *addonsv1alpha1.BreakRequestTemplate) error {
	if !brt.Spec.AutoApprove {
		if brt.Spec.ApprovalCondition != "" {
			return fmt.Errorf("approvalCondition should not be set when autoApprove is false")
		}
		return nil
	}

	if brt.Spec.ApprovalCondition == "" {
		return nil
	}

	if _, err := brt.PrepareCondition(); err != nil {
		return fmt.Errorf("approvalCondition is invalid: %w", err)
	}

	return nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type BreakRequestTemplate.
func (v *BreakRequestTemplateCustomValidator) ValidateUpdate(
	_ context.Context,
	_, newObj runtime.Object,
) (admission.Warnings, error) {
	brt, ok := newObj.(*addonsv1alpha1.BreakRequestTemplate)
	if !ok {
		return nil, fmt.Errorf(
			"expected a BreakRequestTemplate object for the newObj but got %T",
			newObj,
		)
	}
	brtlog.Info("Validation for BreakRequestTemplate upon update", "name", brt.GetName())
	return nil, validate(brt)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type BreakRequestTemplate.
func (v *BreakRequestTemplateCustomValidator) ValidateDelete(
	ctx context.Context,
	obj runtime.Object,
) (admission.Warnings, error) {
	return nil, nil
}
