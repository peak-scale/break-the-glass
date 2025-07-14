// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"
	addonsv1alpha1 "github.com/peak-scale/access-requests/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// MutatingWebhook handles mutating webhook requests.
type AccessRequestMutatingWebhook struct {
	Decoder admission.Decoder
	Client  client.Client
	Log     logr.Logger
}

// Handle processes the admission request and adds a label if necessary.
func (mw *AccessRequestMutatingWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	mw.Log.V(7).Info("Received Request")

	// Decode the object
	ar := &addonsv1alpha1.AccessRequest{}
	if err := mw.Decoder.Decode(req, ar); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// If no subjects were defined we assume, that the user requesting the access request is the owner
	if len(ar.Spec.Subject.Subjects) == 0 {
		ar.Spec.Subject.Subjects = []rbacv1.Subject{
			{
				Kind: "User",
				Name: req.UserInfo.Username,
			},
		}
		mw.Log.V(7).Info("no subjects defined, using request user as subject", "user", req.UserInfo.Username)
	}

	mw.Log.V(7).Info("looking up tenant for namespace", "namespace", ar.GetNamespace())

	tntList := capsulev1beta2.TenantList{}
	if err := mw.Client.List(ctx, &tntList, client.MatchingFields{".status.namespaces": app.GetNamespace()}); err != nil {
		admission.Errored(http.StatusInternalServerError, err)
	}

	mw.Log.V(7).Info("retrieved tenants", "tenants", tntList)

	if len(tntList.Items) == 0 {
		return admission.Allowed("no tenant object")
	}

	tenant := tntList.Items[0]

	mw.Log.V(7).Info("matching tenant", "name", tenant.Name)

	// Only if Tenant is translated
	if !controllerutil.ContainsFinalizer(&tenant, meta.ControllerFinalizer) {
		return admission.Allowed("tenant not translated")
	}

	// Add the label if not present
	if app.Spec.Project == tenant.Name {
		mw.Log.V(7).Info("project already set to tenant")

		return admission.Allowed("tenant already set correctly")
	}

	// Overwrite Project
	app.Spec.Project = tenant.Name

	// Marshal the object back to JSON
	marshaledObj, err := json.Marshal(app)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledObj)
}
