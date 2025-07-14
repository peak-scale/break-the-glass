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

package controller

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	addonsv1alpha1 "github.com/peak-scale/access-requests/api/v1alpha1"
	"github.com/peak-scale/access-requests/internal/metrics"
)

// AccessRequestReconciler reconciles a AccessRequest object
type AccessRequestReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Metrics  metrics.AccessRequestsRecorder
	Recorder record.EventRecorder
	Log      logr.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *AccessRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&addonsv1alpha1.AccessRequest{}).
		Named("accessrequest").
		Complete(r)
}

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *AccessRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("Request.Name", req.Name)

	instance := &addonsv1alpha1.AccessRequest{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {

			r.Metrics.DeleteRequestMetrics(instance)
			log.V(5).Info("Request object not found, could have been deleted after reconcile request")

			return reconcile.Result{}, nil
		}

		r.Log.Error(err, "Error reading the object")

		return reconcile.Result{}, nil
	}

	defer func() {
		r.Metrics.DeleteRequestMetrics(instance)
	}()

	return r.reconcile(
		ctx,
		log,
		instance,
	)
}

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *AccessRequestReconciler) reconcile(
	ctx context.Context,
	log logr.Logger,
	request *addonsv1alpha1.AccessRequest,
) (ctrl.Result, error) {

	switch request.Status.Phase {
	case addonsv1alpha1.RequestPhasePending:
		log.V(5).Info("AccessRequest is pending, waiting for TTL")

	case addonsv1alpha1.RequestPhaseApproved:
		log.V(5).Info("AccessRequest is approved, checking if duration can be started")

	case addonsv1alpha1.RequestPhaseDenied:
		log.V(5).Info("AccessRequest is denied, handling denied state")

	// When the AccessRequest is ongoing
	case addonsv1alpha1.RequestPhaseActive:
		log.V(5).Info("AccessRequest is denied, handling denied state")

	// When the AccessRequest is ongoing
	case addonsv1alpha1.RequestPhaseExpired:
		log.V(5).Info("AccessRequest is expired, Holding expired state until keep date is reached")

	// The case when the AccessRequest is newly created
	default:
		log.V(5).Info("AccessRequest is in requested phase, handling requested state")

	}

	// Always Post Status
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		current := &addonsv1alpha1.AccessRequest{}
		if err := r.Get(ctx, client.ObjectKeyFromObject(request), current); err != nil {
			return fmt.Errorf("failed to refetch instance before update: %w", err)
		}

		current.Status = request.Status

		log.V(7).Info("updating status", "status", current.Status)

		return r.Client.Status().Update(ctx, current)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// When a request is approved, it can be activated immediately or after a certain duration.
func (r *AccessRequestReconciler) transitionRequestActivation(
	ctx context.Context,
	log logr.Logger,
	request *addonsv1alpha1.AccessRequest,
) error {
	if err := request.ActiveRequest(r.Recorder); err != nil {
		return err
	}

	// Reflect Binding

}

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *AccessRequestReconciler) CreatePermissions(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}
