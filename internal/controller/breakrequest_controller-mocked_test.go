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
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	bgv1 "github.com/peak-scale/break-the-glass/api/v1alpha1"
	mc "github.com/peak-scale/break-the-glass/internal/mocks/client"
	gm "go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
)

const resourceName = "test-resource"

var _ = Describe("AccessRequest Controller", func() {

	var (
		ctx      context.Context
		mockCtrl *gm.Controller
		cl       *mc.MockClient
		scl      *mc.MockSubResourceWriter
		s        *runtime.Scheme

		matchBr = gm.AssignableToTypeOf(&bgv1.BreakRequest{})
		matchCm = gm.AssignableToTypeOf(&corev1.ConfigMap{})
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockCtrl = gm.NewController(GinkgoT())
		cl = mc.NewMockClient(mockCtrl)
		scl = mc.NewMockSubResourceWriter(mockCtrl)
		s = scheme.Scheme
		cl.EXPECT().Status().Return(scl).AnyTimes()
		cl.EXPECT().Scheme().Return(s).AnyTimes()
	})
	AfterEach(func() {
		defer mockCtrl.Finish()
	})

	Context("When reconciling a resource", func() {
		var (
			br                   *bgv1.BreakRequest
			controllerReconciler *BreakRequestReconciler
			log                  logr.Logger
		)

		BeforeEach(func() {
			br = &bgv1.BreakRequest{
				ObjectMeta: v1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: bgv1.BreakRequestSpec{
					Items: []runtime.RawExtension{
						{
							Object: &corev1.ConfigMap{
								TypeMeta: v1.TypeMeta{
									Kind: "ConfigMap",
								},
								ObjectMeta: v1.ObjectMeta{
									Name:      "test-configmap",
									Namespace: "default",
								},
							},
						},
					},
				},
			}
			log = ctrl.Log
			controllerReconciler = &BreakRequestReconciler{
				Client:   cl,
				Scheme:   s,
				Recorder: &record.FakeRecorder{},
				Log:      log,
			}
		})

		It("should successfully reconcile the resource that is newly created", func() {
			cl.EXPECT().Get(gm.Any(), gm.Any(), matchBr)
			scl.EXPECT().Update(gm.Any(), matchBr, gm.Any())

			_, err := controllerReconciler.reconcile(ctx, log, br)
			Expect(err).NotTo(HaveOccurred())
			Expect(br.Status.Conditions).To(HaveLen(1))
			Expect(br.Status.Phase).To(Equal(bgv1.RequestPhaseRequested))
		})

		It(
			"should successfully reconcile the resource that is approved but not yet to start",
			func() {
				br.Status.Phase = bgv1.RequestPhaseApproved
				br.Status.Conditions = []v1.Condition{
					{
						LastTransitionTime: v1.Now(),
						Message:            "Access request approved",
						Reason:             "ApprovedByUser",
						Status:             "True",
						Type:               "Approved",
					},
				}
				br.Status.Approved = &bgv1.BreakRequestStatusReviewProperties{
					StartTime: v1.NewTime(time.Now().Add(time.Hour)),
				}

				cl.EXPECT().Get(gm.Any(), gm.Any(), matchBr)
				scl.EXPECT().Update(gm.Any(), matchBr, gm.Any())

				_, err := controllerReconciler.reconcile(ctx, log, br)

				Expect(err).NotTo(HaveOccurred())
				Expect(
					meta.FindStatusCondition(br.Status.Conditions, "Approved"),
				).NotTo(BeNil())
				Expect(br.Status.Phase).To(Equal(bgv1.RequestPhaseApproved))
			},
		)

		It("should successfully reconcile the resource that is approved and ready", func() {
			br.Status.Phase = bgv1.RequestPhaseApproved
			br.Status.Conditions = []v1.Condition{
				{
					LastTransitionTime: v1.Now(),
					Message:            "Access request approved",
					Reason:             "ApprovedByUser",
					Status:             "True",
					Type:               "Approved",
				},
			}
			br.Status.Approved = &bgv1.BreakRequestStatusReviewProperties{
				StartTime: v1.Now(),
			}

			cl.EXPECT().Get(gm.Any(), gm.Any(), matchBr)
			cl.EXPECT().Get(gm.Any(), gm.Any(), matchCm)
			cl.EXPECT().Update(gm.Any(), matchCm, gm.Any())
			scl.EXPECT().Update(gm.Any(), matchBr, gm.Any())

			_, err := controllerReconciler.reconcile(ctx, log, br)

			Expect(err).NotTo(HaveOccurred())

			Expect(meta.FindStatusCondition(br.Status.Conditions, "Approved")).NotTo(BeNil())
			Expect(meta.FindStatusCondition(br.Status.Conditions, "Active")).NotTo(BeNil())
			Expect(br.Status.Phase).To(Equal(bgv1.RequestPhaseActive))

			Expect(br.Status.Approved.Items).To(HaveLen(1))
			cm, ok := br.Status.Approved.Items[0].Object.(*corev1.ConfigMap)
			Expect(ok).To(BeTrue())
			Expect(cm.GetOwnerReferences()).To(HaveLen(1))
		})
	})
})
