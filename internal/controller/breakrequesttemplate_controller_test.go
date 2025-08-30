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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	addonsv1alpha1 "github.com/peak-scale/break-the-glass/api/v1alpha1"
)

var _ = Describe("BreakRequestTemplate Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		breakrequesttemplate := &addonsv1alpha1.BreakRequestTemplate{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind BreakRequestTemplate")
			err := k8sClient.Get(ctx, typeNamespacedName, breakrequesttemplate)
			if err != nil && errors.IsNotFound(err) {
				resource := &addonsv1alpha1.BreakRequestTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: addonsv1alpha1.BreakRequestTemplateSpec{
						Items: items.TemplateItems{
							"cm": items.TemplateItem{
								Item: us(&v1.ConfigMap{
									ObjectMeta: metav1.ObjectMeta{
										Name:      "test-configmap",
										Namespace: "default",
									},
								}),
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &addonsv1alpha1.BreakRequestTemplate{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance BreakRequestTemplate")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &BreakRequestTemplateReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})

func us(obj client.Object) items.Item {
	us, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	return items.Item{Object: us}
}
