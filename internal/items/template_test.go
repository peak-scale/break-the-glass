package items

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("Template", func() {
	Context("RenderTemplate", func() {
		var (
			it     []byte
			params string
		)

		Context("Rendering an item", func() {
			It("Should create the same item if the source has no template params", func() {
				it = tplNestedValue

				res, err := RenderTemplate(it, y2j(params))
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(it))
			})
			It("Should create a valid item if source has no template params", func() {
				it = tplNestedValueParam
				params = "key1: value1"
				res, err := RenderTemplate(it, y2j(params))
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(tplNestedValue))
			})
			It("Should fail if the template is invalid", func() {
				it = y2j(`key1: "{{{.key1}}"`)
				_, err := RenderTemplate(it, y2j(params))
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

var (
	tplNestedValue = y2j(`
key1: value1
key2:
  nestedKey: nestedValue`)
	tplNestedValueParam = y2j(`
key1: "{{.key1}}"
key2:
  nestedKey: nestedValue`)
)

func y2j(in string) []byte {
	m := make(map[string]any)
	err := yaml.Unmarshal([]byte(in), &m)
	if err != nil {
		panic(err)
	}
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}
