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
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mc "github.com/peak-scale/break-the-glass/internal/mocks/client"
	gm "go.uber.org/mock/gomock"
	"sigs.k8s.io/controller-runtime/pkg/client"

	addonsv1alpha1 "github.com/peak-scale/break-the-glass/api/v1alpha1"
)

var _ = Describe("BreakRequest Webhook", func() {
	var (
		br        *addonsv1alpha1.BreakRequest
		validator BreakRequestCustomValidator
		mockCtrl  *gm.Controller
		cl        *mc.MockClient
	)

	BeforeEach(func() {
		mockCtrl = gm.NewController(GinkgoT())
		cl = mc.NewMockClient(mockCtrl)
		br = &addonsv1alpha1.BreakRequest{}
		validator = BreakRequestCustomValidator{
			client: cl,
		}
		Expect(validator).NotTo(BeNil(), "Expected validator to be initialized")
		Expect(br).NotTo(BeNil(), "Expected obj to be initialized")
	})
	AfterEach(func() {
		defer mockCtrl.Finish()
	})

	Context("When creating BreakRequest under Validating Webhook", func() {
		It("Should deny creation if the referenced template is not available", func() {
			By("simulating an invalid creation scenario")
			br.Spec.TemplateName = "foo"
			cl.EXPECT().
				Get(gm.Any(), client.ObjectKey{Name: br.Spec.TemplateName}, gm.Any(), gm.Any()).
				Return(errors.New("not found"))
			Expect(validator.ValidateCreate(ctx, br)).Error().To(HaveOccurred())
		})
		It("Should all creation if the referenced template is available", func() {
			By("simulating an invalid creation scenario")
			br.Spec.TemplateName = "bar"
			cl.EXPECT().
				Get(gm.Any(), client.ObjectKey{Name: br.Spec.TemplateName}, gm.Any(), gm.Any())
			Expect(validator.ValidateCreate(ctx, br)).Error().NotTo(HaveOccurred())
		})
	})
})
