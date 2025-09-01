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
	"errors"
	"fmt"
	"time"

	"github.com/peak-scale/break-the-glass/internal/items"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	addonsv1alpha1 "github.com/peak-scale/break-the-glass/api/v1alpha1"
	"github.com/peak-scale/break-the-glass/internal/meta"
	"github.com/peak-scale/break-the-glass/internal/metrics"
)

type BreakRequestReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Metrics  metrics.BreakRequestsRecorder
	Recorder record.EventRecorder
	Log      logr.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *BreakRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&addonsv1alpha1.BreakRequest{}).
		Named("accessrequest").
		Complete(r)
}

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *BreakRequestReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {
	log := r.Log.WithValues("Request.Name", req.Name).WithValues("Request.Namespace", req.Namespace)

	instance := &addonsv1alpha1.BreakRequest{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {

			r.Metrics.DeleteRequestMetrics(instance)
			log.V(5).
				Info("Request object not found, could have been deleted after reconcile request")

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
func (r *BreakRequestReconciler) reconcile(
	ctx context.Context,
	log logr.Logger,
	request *addonsv1alpha1.BreakRequest,
) (res ctrl.Result, err error) {
	defer func() {
		// Always Post Status
		cerr := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			current := &addonsv1alpha1.BreakRequest{}
			if err := r.Get(ctx, client.ObjectKeyFromObject(request), current); err != nil {
				if apierrors.IsNotFound(err) {
					// if the request is deleted, we cannot find it anymore
					return nil
				}
				return fmt.Errorf("failed to refetch instance before update: %w", err)
			}

			current.Status = request.Status

			log.V(7).Info("updating status", "status", current.Status)

			return r.Client.Status().Update(ctx, current)
		})
		if cerr != nil {
			log.Error(cerr, "failed updating status")
			if err == nil {
				err = cerr
			}
		} else {
			log.V(7).Info("successful update", "status", request.Status)
		}
	}()

	switch request.Status.Phase {
	case addonsv1alpha1.RequestPhasePending:
		log.V(5).Info("BreakRequest is pending, waiting for TTL")

	case addonsv1alpha1.RequestPhaseApproved:
		log.V(5).Info("BreakRequest is approved, checking if duration can be started")

		if request.Status.Approved.StartTime.IsZero() ||
			time.Until(request.Status.Approved.StartTime.Time) <= 0 {
			log.V(5).Info("BreakRequest is approved, activating request")

			// Transition to Active Phase
			if err := r.transitionRequestActivation(ctx, request); err != nil {
				return ctrl.Result{}, fmt.Errorf(
					"failed to activate BreakRequest %s: %w",
					request.Name,
					err,
				)
			}

			log.V(5).Info("BreakRequest activated successfully")
			return ctrl.Result{}, nil
		}

	case addonsv1alpha1.RequestPhaseDenied:
		if err := r.addFinalizer(ctx, log, request); err != nil {
			return ctrl.Result{}, err
		}

		log.V(5).Info("BreakRequest is denied, handling denied state")

		// r.Recorder.Event(request, corev1.EventTypeWarning, "Denied", fmt.Sprintf("Request denied by %s %s", entity.Type, entity.Name))

	case addonsv1alpha1.RequestPhaseActive:
		if err := r.addFinalizer(ctx, log, request); err != nil {
			return ctrl.Result{}, err
		}

		r.Recorder.Event(request, corev1.EventTypeNormal, "Activated", "Request Activated")

		if request.Status.Active != nil {
			if !request.Status.Active.ActiveUntil.IsZero() {
				ts := metav1.Now()
				if ts.After(request.Status.Active.ActiveUntil.Time) {
					r.Recorder.Event(request, corev1.EventTypeNormal, "Expired", "Request Expired")
					return ctrl.Result{}, request.ExpireRequest(nil)
				}

				log.V(5).Info("Requeueing when expiration is due")

				return ctrl.Result{
					RequeueAfter: request.Status.Active.ActiveUntil.Sub(ts.Time),
				}, nil
			}
		}

		return ctrl.Result{}, nil

	// When the BreakRequest has expired
	case addonsv1alpha1.RequestPhaseExpired:
		if request.Status.KeepUntil.Time.IsZero() ||
			time.Until(request.Status.KeepUntil.Time) <= 0 {
			log.V(5).Info("AccessRequest is expired, deleting request")
			return ctrl.Result{}, r.Delete(ctx, request)
		}

		log.V(5).Info(
			"AccessRequest is expired, Holding expired state until keep date (%s) is reached",
			request.Status.KeepUntil.Time,
		)

		if err := r.deleteItems(ctx, request); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: time.Until(request.Status.KeepUntil.Time)}, nil

	// The case when the AccessRequest is newly created
	default:
		log.V(5).Info("AccessRequest is newly created, moving to pending phase")

		if err := request.SetRequested(); err != nil {
			return ctrl.Result{}, err
		}

		r.Recorder.Event(
			request,
			corev1.EventTypeNormal,
			string(request.Status.Phase),
			"Pending Review",
		)
	}

	return ctrl.Result{}, nil
}

// We are adding a finalizer to the BreakRequest to ensure it's not deleted before the request is processed (KeepFor period).
func (r *BreakRequestReconciler) addFinalizer(
	ctx context.Context,
	log logr.Logger,
	request *addonsv1alpha1.BreakRequest,
) error {
	if request.Status.KeepUntil.Time.IsZero() || time.Until(request.Status.KeepUntil.Time) <= 0 {
		return nil
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, request, func() error {
		finalizerName := meta.ControllerFinalizer
		if controllerutil.ContainsFinalizer(request, finalizerName) {
			log.V(5).Info("Finalizer already exists", "name", request.Name)
			return nil
		}

		log.V(5).Info("Adding finalizer to BreakRequest", "name", request.Name)
		controllerutil.AddFinalizer(request, finalizerName)

		return nil
	}); err != nil {
		return fmt.Errorf("failed to add finalizer to BreakRequest %s: %w", request.Name, err)
	}

	return r.Get(ctx, client.ObjectKeyFromObject(request), request)
}

// When a request is approved, it can be activated immediately or after a certain duration.
func (r *BreakRequestReconciler) transitionRequestActivation(
	ctx context.Context,
	request *addonsv1alpha1.BreakRequest,
) error {
	if err := request.ActiveRequest(nil); err != nil {
		return err
	}

	// Reflect Binding
	if err := r.reconcileItems(ctx, request); err != nil {
		return fmt.Errorf("failed to create AccessRequest items %s: %w", request.Name, err)
	}

	return nil
}

// Creates the necessary items resources for the AccessRequest
func (r *BreakRequestReconciler) reconcileItems(
	ctx context.Context,
	request *addonsv1alpha1.BreakRequest,
) (err error) {
	var syncErr error

	brt := &addonsv1alpha1.BreakRequestTemplate{}
	if err := r.Get(ctx, client.ObjectKey{Name: request.Spec.TemplateName}, brt); err != nil {
		return err
	}

	// reset the approved items, only the true approved items should be kept, including the modification done from the operator
	request.Status.Approved.Items = make(items.Items)
	rendered, err := brt.RenderItemsItems(request)
	if err != nil {
		return err
	}

	codecFactory := serializer.NewCodecFactory(r.Client.Scheme())
	for name, raw := range rendered {
		obj := &unstructured.Unstructured{}
		if _, _, decodeErr := codecFactory.UniversalDeserializer().Decode(raw.Raw, nil, obj); decodeErr != nil {
			syncErr = errors.Join(syncErr, decodeErr)
			continue
		}
		obj.SetNamespace(request.Namespace)

		if orerr := controllerutil.SetOwnerReference(request, obj, r.Scheme); orerr != nil {
			syncErr = errors.Join(syncErr, orerr)

			continue
		}

		// append the item to the approved items (use deep copy to avoid using the cluster object)
		request.Status.Approved.Items[name] = &runtime.RawExtension{Object: obj.DeepCopy()}

		// Apply the object to the cluster
		_, err = controllerutil.CreateOrUpdate(ctx, r.Client, obj, func() error {
			labels := obj.GetLabels()
			if labels == nil {
				labels = map[string]string{}
			}

			labels["app.kubernetes.io/managed-by"] = "access-request-controller"
			obj.SetLabels(labels)

			return nil
		})
		if err != nil {
			syncErr = errors.Join(syncErr, err)
		}
	}

	return syncErr
}

// deletes items of the AccessRequest
func (r *BreakRequestReconciler) deleteItems(
	ctx context.Context,
	request *addonsv1alpha1.BreakRequest,
) (err error) {
	var syncErr error

	for _, item := range request.Status.Approved.Items {
		us, err := runtime.DefaultUnstructuredConverter.ToUnstructured(item.Object)
		if err != nil {
			syncErr = errors.Join(syncErr, err)
			continue
		}
		if derr := r.Delete(ctx, &unstructured.Unstructured{Object: us}); derr != nil {
			if !apierrors.IsNotFound(derr) {
				syncErr = errors.Join(syncErr, derr)
				continue
			}
		}
	}

	return syncErr
}
