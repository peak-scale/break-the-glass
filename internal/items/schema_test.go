package items

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.yaml.in/yaml/v3"
)

func y2j(in string) []byte {
	m := make(map[string]any)
	err := yaml.Unmarshal([]byte(in), &m)
	Expect(err).NotTo(HaveOccurred())
	b, err := json.Marshal(m)
	Expect(err).NotTo(HaveOccurred())
	return b
}

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
			`type: object
required: ["key1"]
properties:
  key1:
    type: string
`,
			"key1: value1",
			false,
		), Entry("valid schema and valid params (one allowed extra field)",
			`type: object
required: ["key1"]
properties:
  key1:
    type: string
`,
			`key1: value1
key2: value2`,
			false,
		), Entry("valid schema and invalid params (one additional extra field)",
			`type: object
required: ["key1"]
properties:
  key1:
    type: string
additionalProperties: false
`,
			`key1: value1
key2: value2`,
			true,
		),
		Entry("valid schema but invalid params",
			`type: object
required: ["key1"]
properties:
  key1:
    type: string
`, "key1: 123",
			true,
		),
		Entry("schema missing required field",
			`type: object
required: ["key1"]
properties:
  key1:
    type: string
`, "",
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
