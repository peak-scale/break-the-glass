package items

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Template", func() {
	Context("RenderTemplate", func() {
		var (
			it     Item
			params Params
		)

		BeforeEach(func() {
			it = Item{}
			params = Params{}
		})

		Context("Rendering an item", func() {
			It("Should create the same item if the source has no template params", func() {
				it.Object = map[string]any{
					"key1": "value1",
					"key2": map[string]any{
						"nestedKey": "nestedValue",
					},
				}
				res, err := RenderTemplate(it, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.Object).To(Equal(it.Object))
			})
			It("Should create a valid item if source has no template params", func() {
				it.Object = map[string]any{
					"key1": "{{.key1}}",
					"key2": map[string]any{
						"nestedKey": "nestedValue",
					},
				}
				params = Params{Object: map[string]any{
					"key1": "value1",
				}}
				res, err := RenderTemplate(it, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.Object).To(Equal(map[string]any{
					"key1": "value1",
					"key2": map[string]any{
						"nestedKey": "nestedValue",
					},
				}))
			})
			It("Should fail if the template is invalid", func() {
				it.Object = map[string]any{
					"key1": "{{{.key1}}",
				}
				_, err := RenderTemplate(it, params)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
