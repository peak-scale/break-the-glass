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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/peak-scale/break-the-glass/internal/items"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/peak-scale/break-the-glass/api/v1alpha1"
)

var _ = Describe("BreakRequest Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"
		const templateName = "test-template"

		ctx := context.Background()

		nnBr := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		nnBrt := types.NamespacedName{
			Name: templateName,
		}
		br := &v1alpha1.BreakRequest{}
		brt := &v1alpha1.BreakRequestTemplate{}
		var controllerReconciler *BreakRequestReconciler

		BeforeEach(func() {
			By("creating the custom resource for the Kind BreakRequest")
			err := k8sClient.Get(ctx, nnBrt, brt)
			if err != nil && errors.IsNotFound(err) {
				resource := &v1alpha1.BreakRequestTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name: templateName,
					},
					Spec: v1alpha1.BreakRequestTemplateSpec{
						Items: items.TemplateItems{
							templateName: {
								ManifestTemplate: runtime.RawExtension{Raw: []byte(`{
  "kind": "ConfigMap",
  "metadata": {
    "name": "test-configmap"
  },
  "data": {
    "test": "{{.key1}}"
  }
}`)},
								ParamSchema: runtime.RawExtension{
									Raw: []byte(`{"type": "string"}`),
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
			err = k8sClient.Get(ctx, nnBr, br)
			if err != nil && errors.IsNotFound(err) {
				resource := &v1alpha1.BreakRequest{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: v1alpha1.BreakRequestSpec{
						TemplateName: templateName,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
			controllerReconciler = &BreakRequestReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: &record.FakeRecorder{},
				Log:      ctrl.Log,
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &v1alpha1.BreakRequest{}
			err := k8sClient.Get(ctx, nnBr, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance BreakRequest")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: nnBr,
			})
			Expect(err).NotTo(HaveOccurred())
			resource := &v1alpha1.BreakRequest{}
			err = k8sClient.Get(ctx, nnBr, resource)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
