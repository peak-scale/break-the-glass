package items

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OpenAPI Schema", func() {
	DescribeTable("Validate",
		func(schemaJSON map[string]any, params map[string]any, expectErr bool) {
			paramSchema := ParamSchema{
				Object: schemaJSON,
			}
			p := Params{
				Object: params,
			}

			err := Validate(paramSchema, p)
			if expectErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		},
		Entry("valid schema and valid params",
			map[string]any{
				"type":     "object",
				"required": []string{"key1"},
				"properties": map[string]any{
					"key1": map[string]any{
						"type": "string",
					},
				},
			},
			map[string]any{
				"key1": "value1",
			},
			false,
		), Entry("valid schema and valid params (one allowed extra field)",
			map[string]any{
				"type":     "object",
				"required": []string{"key1"},
				"properties": map[string]any{
					"key1": map[string]any{
						"type": "string",
					},
				},
			},
			map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			false,
		), Entry("valid schema and invalid params (one additional extra field)",
			map[string]any{
				"type":     "object",
				"required": []string{"key1"},
				"properties": map[string]any{
					"key1": map[string]any{
						"type": "string",
					},
				},
				"additionalProperties": false,
			},
			map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			true,
		),
		Entry("valid schema but invalid params",
			map[string]any{
				"type":     "object",
				"required": []string{"key1"},
				"properties": map[string]any{
					"key1": map[string]any{
						"type": "string",
					},
				},
			},
			map[string]any{
				"key1": 123,
			},
			true,
		),
		Entry("schema missing required field",
			map[string]any{
				"type":     "object",
				"required": []string{"key1"},
				"properties": map[string]any{
					"key1": map[string]any{
						"type": "string",
					},
				},
			},
			map[string]any{},
			true,
		),
		Entry("invalid schema JSON",
			map[string]any{
				"type": nil,
			},
			map[string]any{
				"key1": "value1",
			},
			true,
		),
		Entry("empty schema and params",
			map[string]any{},
			map[string]any{},
			false,
		),
	)
})
