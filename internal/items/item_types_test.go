package items

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Params", func() {
	Context("DeepCopyInto", func() {
		type testCase struct {
			name       string
			input      Params
			expected   Params
			shouldFail bool
		}

		DescribeTable("DeepCopyInto",
			func(tc testCase) {
				var result Params
				tc.input.DeepCopyInto(&result)

				resultBytes, err := json.Marshal(result.Object)
				Expect(err).NotTo(HaveOccurred())

				expectedBytes, err := json.Marshal(tc.expected.Object)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(resultBytes)).To(Equal(string(expectedBytes)))
			},
			Entry("valid deep copy", testCase{
				name: "valid deep copy",
				input: Params{
					Object: map[string]any{
						"key1": "value1",
						"key2": map[string]any{
							"nestedKey": "nestedValue",
						},
					},
				},
				expected: Params{
					Object: map[string]any{
						"key1": "value1",
						"key2": map[string]any{
							"nestedKey": "nestedValue",
						},
					},
				},
				shouldFail: false,
			}),
			Entry("empty input copy", testCase{
				name:       "empty input copy",
				input:      Params{Object: map[string]any{}},
				expected:   Params{Object: map[string]any{}},
				shouldFail: false,
			}),
			Entry("nil input", testCase{
				name: "nil input",
				input: Params{
					Object: nil,
				},
				expected: Params{
					Object: nil,
				},
				shouldFail: false,
			}),
		)
	})
})
