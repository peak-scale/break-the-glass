// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"
	addonsv1alpha1 "github.com/peak-scale/break-the-glass/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	ar := &addonsv1alpha1.BreakRequest{}
	if err := mw.Decoder.Decode(req, ar); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	ar.Spec.Requestor = addonsv1alpha1.AccessEntity{
		Name: req.UserInfo.Username,
		Type: addonsv1alpha1.AccessEntityTypeUser,
	}

	// Marshal the object back to JSON
	marshaledObj, err := json.Marshal(ar)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledObj)
}

// MutatingWebhook handles mutating webhook requests.
type AccessRequestValidatingWebhook struct {
	Decoder admission.Decoder
	Client  client.Client
	Log     logr.Logger
}

// Handle processes the admission request and adds a label if necessary.
func (mw *AccessRequestValidatingWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	mw.Log.V(7).Info("Received Request")

	// Decode the object
	ar := &addonsv1alpha1.BreakRequest{}
	if err := mw.Decoder.Decode(req, ar); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	return admission.Allowed("spec update allowed")

}
