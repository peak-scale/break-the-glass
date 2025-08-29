package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExtendedDuration", func() {
	DescribeTable(
		"IsApproved",
		func(spec BreakRequestTemplateSpec, br BreakRequest, approved, expectError bool) {
			brt := &BreakRequestTemplate{Spec: spec}
			result, err := brt.IsApproved(&br)
			if expectError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(approved))
			}
		},
		Entry(
			"Not approved if no auto approval case 1",
			BreakRequestTemplateSpec{AutoApprove: false},
			BreakRequest{},
			false,
			false,
		),
		Entry(
			"Approved if auto approval and no condition",
			BreakRequestTemplateSpec{AutoApprove: true, ApprovalCondition: ""},
			BreakRequest{},
			true,
			false,
		),
		Entry(
			"Reason is correct",
			BreakRequestTemplateSpec{
				AutoApprove:       true,
				ApprovalCondition: "request.spec.reason == 'test'",
			},
			BreakRequest{Spec: BreakRequestSpec{Reason: "test"}},
			true,
			false,
		),
		Entry(
			"Reason is incorrect",
			BreakRequestTemplateSpec{
				AutoApprove:       true,
				ApprovalCondition: "request.spec.reason == 'test'",
			},
			BreakRequest{Spec: BreakRequestSpec{Reason: "TEST"}},
			false,
			false,
		),
	)
})
