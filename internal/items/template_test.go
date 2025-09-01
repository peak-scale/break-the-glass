package items

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Template", func() {
	Context("RenderTemplate", func() {
		var (
			it     string
			params string
		)

		Context("Rendering an item", func() {
			It("Should create the same item if the source has no template params", func() {
				it = `key1: value1
key2:
  nestedKey: nestedValue`

				res, err := RenderTemplate(y2j(it), y2j(params))
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(y2j(it)))
			})
			It("Should create a valid item if source has no template params", func() {
				it = `key1: "{{.key1}}"
key2:
  nestedKey: nestedValue`
				params = "key1: value1"
				res, err := RenderTemplate(y2j(it), y2j(params))
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(y2j(`key1: value1
key2:
  nestedKey: nestedValue`)))
			})
			It("Should fail if the template is invalid", func() {
				it = `key1: "{{{.key1}}"`
				_, err := RenderTemplate(y2j(it), y2j(params))
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
