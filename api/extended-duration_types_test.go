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

package api

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const kubeBuilderType = "+kubebuilder:validation:Type="

var _ = Describe("ExtendedDuration", func() {

	Context("UnmarshalJSON", func() {
		DescribeTable("should unmarshal JSON strings correctly",
			func(input string, expectErr bool, expected ExtendedDuration) {
				var d ExtendedDuration
				err := d.UnmarshalJSON([]byte(input))
				if expectErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(d).To(Equal(expected))
				}
			},
			Entry("valid duration", `"1h30m"`, false, ExtendedDuration(time.Hour+30*time.Minute)),
			Entry("valid zero duration", `"0s"`, false, ExtendedDuration(0)),
			Entry("invalid format", `"not-a-duration"`, true, ExtendedDuration(0)),
			Entry("empty input", `""`, true, ExtendedDuration(0)),
		)
	})

	Context("String", func() {
		DescribeTable("should convert to string correctly",
			func(input ExtendedDuration, expected string) {
				result := input.String()
				Expect(result).To(Equal(expected))
			},
			Entry("one hour", ExtendedDuration(time.Hour), "1h"),
			Entry("hour and minutes", ExtendedDuration(time.Hour+30*time.Minute), "1h30m"),
			Entry("zero duration", ExtendedDuration(0), "0s"),
		)
	})

	Context("MarshalJSON", func() {
		DescribeTable("should marshal to JSON correctly",
			func(input ExtendedDuration, expectErr bool, expected string) {
				result, err := input.MarshalJSON()
				if expectErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(string(result)).To(Equal(expected))
				}
			},
			Entry("one hour", ExtendedDuration(time.Hour), false, `"1h"`),
			Entry("hour and minutes", ExtendedDuration(time.Hour+30*time.Minute), false, `"1h30m"`),
			Entry("zero duration", ExtendedDuration(0), false, `"0s"`),
		)
	})

	Context("ToUnstructured", func() {
		DescribeTable("should convert to unstructured correctly",
			func(input ExtendedDuration, expected string) {
				result := input.ToUnstructured()
				Expect(result).To(Equal(expected))
			},
			Entry("one hour", ExtendedDuration(time.Hour), "1h"),
			Entry("hour and minutes", ExtendedDuration(time.Hour+30*time.Minute), "1h30m"),
			Entry("zero duration", ExtendedDuration(0), "0s"),
		)
	})

	Context("OpenAPISchemaType", func() {
		It("should return correct schema type", func() {
			var d ExtendedDuration
			Expect(d.OpenAPISchemaType()).To(Equal([]string{"string"}))
		})
		It("should verify the schema type matches the kubebuilder comment", func() {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(
				fset,
				"extended-duration_types.go",
				nil,
				parser.ParseComments,
			)
			if err != nil {
				Expect(err).NotTo(HaveOccurred())
			}

			var d ExtendedDuration
			schemaType := findKubeBuilderComment(file)
			Expect(schemaType).To(HaveLen(1))
			Expect(schemaType[0]).To(Equal(d.OpenAPISchemaType()[0]))

		})
	})

	Context("OpenAPISchemaFormat", func() {
		It("should return correct schema format", func() {
			var d ExtendedDuration
			Expect(d.OpenAPISchemaFormat()).To(Equal(""))
		})
	})
})

func findKubeBuilderComment(file *ast.File) []string {
	var schemaType []string
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if strings.Contains(c.Text, kubeBuilderType) {
				schemaType = append(
					schemaType,
					strings.Split(c.Text, kubeBuilderType)[1],
				)
			}
		}
	}
	return schemaType
}
