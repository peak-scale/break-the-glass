package conditions

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/peak-scale/break-the-glass/api/v1alpha1"
)

var _ = Describe("Condition", func() {
	DescribeTable(
		"IsApproved",
		func(spec v1alpha1.BreakRequestTemplateSpec, br v1alpha1.BreakRequest, approved, expectError bool) {
			brt := &v1alpha1.BreakRequestTemplate{Spec: spec}
			result, err := IsApproved(brt, &br)
			if expectError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(approved))
			}
		},
		Entry(
			"Not approved if no auto approval case 1",
			v1alpha1.BreakRequestTemplateSpec{AutoApprove: false},
			v1alpha1.BreakRequest{},
			false,
			false,
		),
		Entry(
			"Approved if auto approval and no condition",
			v1alpha1.BreakRequestTemplateSpec{AutoApprove: true, ApprovalCondition: ""},
			v1alpha1.BreakRequest{},
			true,
			false,
		),
		Entry(
			"Reason is correct",
			v1alpha1.BreakRequestTemplateSpec{
				AutoApprove:       true,
				ApprovalCondition: "request.spec.reason == 'test'",
			},
			v1alpha1.BreakRequest{Spec: v1alpha1.BreakRequestSpec{Reason: "test"}},
			true,
			false,
		),
		Entry(
			"Reason is incorrect",
			v1alpha1.BreakRequestTemplateSpec{
				AutoApprove:       true,
				ApprovalCondition: "request.spec.reason == 'test'",
			},
			v1alpha1.BreakRequest{Spec: v1alpha1.BreakRequestSpec{Reason: "TEST"}},
			false,
			false,
		),
	)
})
