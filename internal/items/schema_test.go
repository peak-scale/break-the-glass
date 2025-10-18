package items

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OpenAPI Schema", func() {
	DescribeTable("Validate",
		func(schemaJSON string, params string, expectErr bool) {
			err := Validate(y2j(schemaJSON), y2j(params))
			if expectErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		},
		Entry("valid schema and valid params",
			schemaString,
			"key1: value1",
			false,
		), Entry("valid schema and valid params (one allowed extra field)",
			schemaString,
			paramKey1Key2,
			false,
		), Entry("valid schema and invalid params (one additional extra field)",
			schemaStringNoAdditionalProperties,
			paramKey1Key2,
			true,
		),
		Entry("valid schema but invalid params",
			schemaString, "key1: 123",
			true,
		),
		Entry("schema missing required field",
			schemaString, "",
			true,
		),
		Entry("invalid schema JSON",
			"type:",
			"key1: value1",
			true,
		),
		Entry("empty schema and params",
			"",
			"",
			false,
		),
	)
})

var (
	schemaString = `
type: object
required: ["key1"]
properties:
  key1:
    type: string
`
	schemaStringNoAdditionalProperties = schemaString + `
additionalProperties: false`

	paramKey1Key2 = `
key1: value1
key2: value2`
)
