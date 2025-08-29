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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	addonsv1alpha1 "github.com/peak-scale/break-the-glass/api/v1alpha1"
)

var _ = Describe("BreakRequestTemplate Webhook", func() {
	var (
		obj       *addonsv1alpha1.BreakRequestTemplate
		validator BreakRequestTemplateCustomValidator
	)

	BeforeEach(func() {
		obj = &addonsv1alpha1.BreakRequestTemplate{}
		validator = BreakRequestTemplateCustomValidator{}
		Expect(validator).NotTo(BeNil(), "Expected validator to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
	})

	Context("When creating or updating BreakRequestTemplate under Validating Webhook", func() {
		Context("When auto approval is enabled an condition is not empty", func() {
			BeforeEach(func() {
				obj.Spec.AutoApprove = false
				obj.Spec.ApprovalCondition = "foo"
			})
			It("Should deny creation", func() {
				By("simulating an invalid creation scenario")
				Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
			})
			It("Should deny update", func() {
				By("simulating an invalid update scenario")
				Expect(validator.ValidateUpdate(ctx, nil, obj)).Error().To(HaveOccurred())
			})
		})

		Context("When auto approval is enabled an condition is empty", func() {
			BeforeEach(func() {
				obj.Spec.AutoApprove = true
			})
			It("Should allow creation", func() {
				By("simulating an valid creation scenario")
				Expect(validator.ValidateCreate(ctx, obj)).Error().NotTo(HaveOccurred())
			})
			It("Should allow update", func() {
				By("simulating an valid update scenario")
				Expect(validator.ValidateUpdate(ctx, nil, obj)).Error().NotTo(HaveOccurred())
			})
		})

		Context("When auto approval is enabled an condition is invalid", func() {
			BeforeEach(func() {
				obj.Spec.AutoApprove = true
				obj.Spec.ApprovalCondition = "foo.spec.reason == 'test'"
			})
			It("Should deny creation", func() {
				By("simulating an invalid creation scenario")
				Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
			})
			It("Should deny update", func() {
				By("simulating an invalid update scenario")
				Expect(validator.ValidateUpdate(ctx, nil, obj)).Error().To(HaveOccurred())
			})
		})

		Context("When auto approval is enabled an condition is valid", func() {
			BeforeEach(func() {
				obj.Spec.AutoApprove = true
				obj.Spec.ApprovalCondition = "request.spec.reason == 'test'"
			})
			It("Should allow creation", func() {
				By("simulating an valid creation scenario")
				Expect(validator.ValidateCreate(ctx, obj)).Error().NotTo(HaveOccurred())
			})
			It("Should allow update", func() {
				By("simulating an valid update scenario")
				Expect(validator.ValidateUpdate(ctx, nil, obj)).Error().NotTo(HaveOccurred())
			})
		})
	})
})
